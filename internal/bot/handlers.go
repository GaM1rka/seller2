package bot

import (
	"context"
	"fmt"
	"log"
	"seller2/internal/store"
	"strings"
	"time"

	"seller2/config"
	"seller2/internal/data"
	"seller2/internal/messages"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	cbNichePrefix = "niche:" // niche:<key>
	cbRefsPrefix  = "refs:"  // refs:<key>
	cbMenu        = "menu"   // меню
	cbHowPrefix   = "how:"   // how:<key>
	lessonChatID  = int64(-1003212181419)
	lessonMsgID   = 28
)

type Handler struct {
	bot   *Bot
	cfg   config.Config
	store *store.RedisStore
}

func NewHandler(b *Bot, cfg config.Config) *Handler {
	return &Handler{bot: b, cfg: cfg}
}
func NewHandlerWithStore(b *Bot, cfg config.Config, s *store.RedisStore) *Handler {
	return &Handler{bot: b, cfg: cfg, store: s}
}

func (h *Handler) Start() {
	for update := range h.bot.Updates() {
		switch {
		case update.Message != nil:
			h.onMessage(update.Message)
		case update.CallbackQuery != nil:
			h.onCallback(update.CallbackQuery)
		}
	}
}

// -------- message flow ----------

func (h *Handler) onMessage(m *tgbotapi.Message) {
	if m.IsCommand() && m.Command() == "start" {
		// Первый вход — покажем приветствие
		h.sendWelcome(m.Chat.ID)
		return
	}
	if m.Text != "" && strings.EqualFold(m.Text, "start") {
		h.sendWelcome(m.Chat.ID)
		return
	}
	// По умолчанию просто откроем меню
	h.sendMenuOnly(m.Chat.ID)
}

func (h *Handler) onCallback(q *tgbotapi.CallbackQuery) {
	dataStr := q.Data
	switch {
	case dataStr == cbMenu:
		// По кнопке «меню» всегда отправляем НОВОЕ сообщение с выбором ниш
		h.sendMenuOnly(q.Message.Chat.ID)

	case strings.HasPrefix(dataStr, cbNichePrefix):
		key := strings.TrimPrefix(dataStr, cbNichePrefix)
		h.sendNicheGif(q.Message.Chat.ID, key)

	case strings.HasPrefix(dataStr, cbRefsPrefix):
		key := strings.TrimPrefix(dataStr, cbRefsPrefix)
		h.sendRefsFlow(q.Message.Chat.ID, key)

	case strings.HasPrefix(dataStr, cbHowPrefix):
		key := strings.TrimPrefix(dataStr, cbHowPrefix)
		h.sendHowFlow(q.Message.Chat.ID, key)
	}
	_ = h.answer(q)
}

func (h *Handler) answer(q *tgbotapi.CallbackQuery) error {
	cfg := tgbotapi.NewCallback(q.ID, "")
	_, err := h.bot.API.Request(cfg)
	return err
}

// -------- UI builders ----------

func (h *Handler) menuKeyboard() tgbotapi.InlineKeyboardMarkup {
	// Кнопки в столбик, как на скрине
	rows := [][]tgbotapi.InlineKeyboardButton{}
	for _, visible := range data.NicheOrder {
		key := data.NameToKey[visible]
		btn := tgbotapi.NewInlineKeyboardButtonData(visible, cbNichePrefix+key)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (h *Handler) menuMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = h.menuKeyboard()
	h.mustSend(msg)
}

func (h *Handler) sendWelcome(chatID int64) {
	// Приветствие при /start
	h.menuMessage(chatID, messages.Welcome)
}

// короткая версия меню — именно её шлём по кнопке «меню»
func (h *Handler) sendMenuOnly(chatID int64) {
	h.menuMessage(chatID, "Выбери нишу ниже 👇")
}

func (h *Handler) twoButtonsMenuRefs(key string) tgbotapi.InlineKeyboardMarkup {
	btnMenu := tgbotapi.NewInlineKeyboardButtonData("меню", cbMenu)
	btnRefs := tgbotapi.NewInlineKeyboardButtonData("референсы", cbRefsPrefix+key)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btnMenu, btnRefs),
	)
}

func (h *Handler) twoButtonsHowMenu(key string) tgbotapi.InlineKeyboardMarkup {
	btnHow := tgbotapi.NewInlineKeyboardButtonData("🎥 Показать, как это делается", cbHowPrefix+key)
	btnMenu := tgbotapi.NewInlineKeyboardButtonData("меню", cbMenu)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btnHow),
		tgbotapi.NewInlineKeyboardRow(btnMenu),
	)
}

func (h *Handler) buyKeyboard() tgbotapi.InlineKeyboardMarkup {
	btn := tgbotapi.NewInlineKeyboardButtonURL("«Взять доступ»", h.cfg.TributeURL)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btn),
	)
}

// -------- steps ----------

func (h *Handler) sendNicheGif(chatID int64, key string) {
	n, ok := data.Niches[key]
	if !ok {
		h.sendMenuOnly(chatID)
		return
	}

	// Копируем заранее загруженный в канал пост-гифку
	caption := messages.NicheGifCaption(n.Emoji, n.CaptionWord)
	copy := tgbotapi.NewCopyMessage(chatID, n.Gif.FromChatID, n.Gif.MessageID)
	copy.Caption = caption
	copy.ReplyMarkup = h.twoButtonsMenuRefs(key)

	if _, err := h.bot.API.Request(copy); err != nil {
		log.Printf("copy gif error: %v", err)
		h.menuMessage(chatID, "Не удалось отправить примеры. Проверь доступ бота к каналу-источнику.")
		return
	}
}

func (h *Handler) sendRefsFlow(chatID int64, key string) {
	n, ok := data.Niches[key]
	if !ok {
		h.sendMenuOnly(chatID)
		return
	}

	// Проверка доступа к источнику (однократно на ключ)
	if len(n.Posts) > 0 {
		from := n.Posts[0].FromChatID
		if err := h.checkSourceAccess(from); err != nil {
			log.Printf("no access to source %d: %v", from, err)
			h.menuMessage(chatID, "Не могу получить референсы (нет доступа к источнику). Проверь, что бот добавлен в канал и история доступна.")
			return
		}
	}

	// Копируем 3 поста
	for _, p := range n.Posts {
		copy := tgbotapi.NewCopyMessage(chatID, p.FromChatID, p.MessageID)
		if _, err := h.bot.API.Request(copy); err != nil {
			log.Printf("copy error chat=%d msg=%d: %v", p.FromChatID, p.MessageID, err)
		}
		time.Sleep(150 * time.Millisecond) // маленький троттлинг
	}

	// Через минуту — CTA
	time.AfterFunc(time.Minute, func() {
		msg := tgbotapi.NewMessage(chatID, messages.AfterRefs)
		msg.ReplyMarkup = h.twoButtonsHowMenu(key)
		h.mustSend(msg)
	})
}

func (h *Handler) checkSourceAccess(fromChatID int64) error {
	_, err := h.bot.API.GetChat(tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{ChatID: fromChatID},
	})
	return err
}

func (h *Handler) sendHowFlow(chatID int64, key string) {
	// форвардим урок
	fw := tgbotapi.NewForward(chatID, lessonChatID, lessonMsgID)
	msg := h.mustSend(fw)

	// планируем удаление через 24 часа в Redis
	if h.store != nil {
		_ = h.store.ScheduleDeletion(context.Background(), chatID, msg.MessageID, time.Now().Add(time.Minute))
	} else {
		// fallback: локальный таймер (если нет Redis)
		time.AfterFunc(time.Minute, func() {
			h.deleteLessonAndOffer(chatID, msg.MessageID)
		})
	}
}

func (h *Handler) deleteLessonAndOffer(chatID int64, msgID int) {
	// удалить урок
	del := tgbotapi.DeleteMessageConfig{ChatID: chatID, MessageID: msgID}
	if _, err := h.bot.API.Request(del); err != nil {
		log.Println("delete lesson:", err)
	}
	// оффер
	txt := fmt.Sprintf(messages.Sales, h.cfg.PriceText)
	m := tgbotapi.NewMessage(chatID, txt)
	m.ReplyMarkup = h.buyKeyboard()
	h.mustSend(m)
}

func (h *Handler) RunDeletionScheduler(ctx context.Context) {
	if h.store == nil {
		return
	}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tasks, err := h.store.FetchDue(ctx, time.Now(), 100)
			if err != nil {
				if err.Error() != "redis: nil" {
					log.Println("scheduler fetch:", err)
				}
				continue
			}
			for _, t := range tasks {
				h.deleteLessonAndOffer(t.ChatID, t.MsgID)
				// если хочется ретраев — оборачиваем try/catch и при ошибке перекидываем задачу назад в Redis на +1 минуту
			}
		}
	}
}

// -------- helpers ----------

func (h *Handler) mustSend(c tgbotapi.Chattable) tgbotapi.Message {
	m, err := h.bot.API.Send(c)
	if err != nil {
		log.Println("send error:", err)
	}
	return m
}

func (h *Handler) mustRequest(c tgbotapi.Chattable) {
	if _, err := h.bot.API.Request(c); err != nil {
		log.Println("request error:", err)
	}
}

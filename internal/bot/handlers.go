package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"seller2/internal/store"
	"strconv"
	"strings"
	"time"

	"seller2/config"
	"seller2/internal/data"
	"seller2/internal/messages"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	cbNichePrefix = "niche:"    // niche:<key>
	cbRefsPrefix  = "refs:"     // refs:<key>
	cbMoreRefs    = "morerefs:" // morerefs:<key>:<index>
	cbMenu        = "menu"      // меню
	cbHowPrefix   = "how:"      // how:<key>
	lessonChatID  = int64(-1003212181419)
	lessonMsgID   = 34
	salesChatID   = int64(-1003212181419) // тот же канал
	salesMsgID    = 41                    // ID продающего сообщения
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
	h.menuMessage(chatID, "выбери нишу ниже 👇")
}

func (h *Handler) oneButtonMenu() tgbotapi.InlineKeyboardMarkup {
	btnMenu := tgbotapi.NewInlineKeyboardButtonData("меню", cbMenu)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btnMenu),
	)
}

func (h *Handler) twoButtonsHowMenu(key string) tgbotapi.InlineKeyboardMarkup {
	btnHow := tgbotapi.NewInlineKeyboardButtonData("🎥 показать, как это делается", cbHowPrefix+key)
	btnMenu := tgbotapi.NewInlineKeyboardButtonData("меню", cbMenu)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btnHow),
		tgbotapi.NewInlineKeyboardRow(btnMenu),
	)
}

func (h *Handler) buyKeyboard() tgbotapi.InlineKeyboardMarkup {
	btn := tgbotapi.NewInlineKeyboardButtonURL("«взять доступ»", h.cfg.TributeURL)
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btn),
	)
}

// -------- steps ----------

func (h *Handler) sendNicheFlow(chatID int64, key string) {
	n, ok := data.Niches[key]
	if !ok {
		h.sendMenuOnly(chatID)
		return
	}

	// 1) Гиф-пост с кастомной подписью и ТОЛЬКО «меню»
	caption := messages.NicheGifCaption(n.Emoji, n.CaptionWord)
	copy := tgbotapi.NewCopyMessage(chatID, n.Gif.FromChatID, n.Gif.MessageID)
	copy.Caption = caption
	copy.ReplyMarkup = h.oneButtonMenu()
	if _, err := h.bot.API.Request(copy); err != nil {
		log.Printf("copy gif error: %v", err)
		h.menuMessage(chatID, "не удалось отправить примеры. проверь доступ бота к каналу-источнику.")
		return
	}

	// 2) Отправляем первый референс с кнопкой "еще референс"
	h.sendNextRef(chatID, key, 0)
}

// sendNextRef отправляет следующий референс (по индексу)
func (h *Handler) sendNextRef(chatID int64, key string, index int) {
	n, ok := data.Niches[key]
	if !ok || index >= len(n.Posts) {
		return
	}

	// Отправляем текущий референс
	p := n.Posts[index]
	copy := tgbotapi.NewCopyMessage(chatID, p.FromChatID, p.MessageID)

	// Добавляем клавиатуру с кнопками
	copy.ReplyMarkup = h.refsKeyboard(key, index, len(n.Posts))

	if _, err := h.bot.API.Request(copy); err != nil {
		log.Printf("copy ref error chat=%d msg=%d: %v", p.FromChatID, p.MessageID, err)
		// Если ошибка, все равно предлагаем следующий
		if index+1 < len(n.Posts) {
			h.sendNextRef(chatID, key, index+1)
		}
		return
	}

	time.Sleep(150 * time.Millisecond)
}

// refsKeyboard создает клавиатуру для референсов
func (h *Handler) refsKeyboard(key string, currentIndex int, total int) tgbotapi.InlineKeyboardMarkup {
	rows := [][]tgbotapi.InlineKeyboardButton{}

	// Если есть еще референсы, показываем кнопку "еще референс"
	if currentIndex+1 < total {
		btnMore := tgbotapi.NewInlineKeyboardButtonData(
			"еще референс",
			fmt.Sprintf("%s%s:%d", cbMoreRefs, key, currentIndex+1),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btnMore))
	}

	// Всегда показываем кнопку "меню"
	btnMenu := tgbotapi.NewInlineKeyboardButtonData("меню", cbMenu)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(btnMenu))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// Обработка callback'ов (добавляем в switch)
func (h *Handler) onCallback(q *tgbotapi.CallbackQuery) {
	dataStr := q.Data
	switch {
	case dataStr == cbMenu:
		h.sendMenuOnly(q.Message.Chat.ID)

	case strings.HasPrefix(dataStr, cbNichePrefix):
		key := strings.TrimPrefix(dataStr, cbNichePrefix)
		h.sendNicheFlow(q.Message.Chat.ID, key)

	case strings.HasPrefix(dataStr, cbRefsPrefix):
		key := strings.TrimPrefix(dataStr, cbRefsPrefix)
		h.sendRefsFlow(q.Message.Chat.ID, key)

	case strings.HasPrefix(dataStr, cbMoreRefs):
		// Обработка "еще референс": morerefs:<key>:<index>
		parts := strings.Split(strings.TrimPrefix(dataStr, cbMoreRefs), ":")
		if len(parts) == 2 {
			key := parts[0]
			index, err := strconv.Atoi(parts[1])
			if err == nil {
				h.sendNextRef(q.Message.Chat.ID, key, index)

				// Если это был последний референс, отправляем финальное сообщение
				n, ok := data.Niches[key]
				if ok && index == len(n.Posts)-1 {
					time.AfterFunc(500*time.Millisecond, func() {
						h.sendFinalMessage(q.Message.Chat.ID, key)
					})
				}
			}
		}

	case strings.HasPrefix(dataStr, cbHowPrefix):
		key := strings.TrimPrefix(dataStr, cbHowPrefix)
		h.sendHowFlow(q.Message.Chat.ID, key)
	}
	_ = h.answer(q)
}

// sendFinalMessage отправляет финальное сообщение после всех референсов
func (h *Handler) sendFinalMessage(chatID int64, key string) {
	msg := tgbotapi.NewMessage(chatID, messages.AfterRefs)
	msg.ReplyMarkup = h.twoButtonsHowMenu(key)
	h.mustSend(msg)
}

// Обновляем sendRefsFlow для использования новой логики
func (h *Handler) sendRefsFlow(chatID int64, key string) {
	// Проверка доступа к источнику
	n, ok := data.Niches[key]
	if !ok {
		h.sendMenuOnly(chatID)
		return
	}

	if len(n.Posts) > 0 {
		from := n.Posts[0].FromChatID
		if err := h.checkSourceAccess(from); err != nil {
			log.Printf("no access to source %d: %v", from, err)
			h.menuMessage(chatID, "не могу получить референсы (нет доступа к источнику). проверь, что бот добавлен в канал и история доступна.")
			return
		}
	}

	// Запускаем отправку первого референса
	h.sendNextRef(chatID, key, 0)
}

func (h *Handler) checkSourceAccess(fromChatID int64) error {
	_, err := h.bot.API.GetChat(tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{ChatID: fromChatID},
	})
	return err
}

func (h *Handler) sendHowFlow(chatID int64, key string) {
	// копируем урок без "forwarded from"
	copy := tgbotapi.NewCopyMessage(chatID, lessonChatID, lessonMsgID)
	resp, err := h.bot.API.Request(copy)
	if err != nil {
		log.Println("copy lesson error:", err)
		return
	}
	var mid tgbotapi.MessageID
	if err := json.Unmarshal(resp.Result, &mid); err != nil {
		log.Println("decode message_id error:", err)
		return
	}
	log.Printf("lesson copied: chat=%d msg_id=%d", chatID, mid.MessageID)

	// расписание
	offerAt := time.Now().Add(time.Minute)      // оффер через 15 минут
	deleteAt := time.Now().Add(2 * time.Minute) // удаление через 24 часа

	if h.store != nil {
		if err := h.store.ScheduleOffer(context.Background(), chatID, mid.MessageID, offerAt); err != nil {
			log.Println("ScheduleOffer error, fallback to AfterFunc:", err)
			time.AfterFunc(time.Until(offerAt), func() { h.sendOffer(chatID) })
		}
		if err := h.store.ScheduleDeletion(context.Background(), chatID, mid.MessageID, deleteAt); err != nil {
			log.Println("ScheduleDeletion error, fallback to AfterFunc:", err)
			time.AfterFunc(time.Until(deleteAt), func() { h.deleteLesson(chatID, mid.MessageID) })
		}
	} else {
		// фоллбэки без Redis
		time.AfterFunc(time.Until(offerAt), func() { h.sendOffer(chatID) })
		time.AfterFunc(time.Until(deleteAt), func() { h.deleteLesson(chatID, mid.MessageID) })
	}
}

func (h *Handler) deleteLesson(chatID int64, msgID int) {
	del := tgbotapi.DeleteMessageConfig{ChatID: chatID, MessageID: msgID}
	if _, err := h.bot.API.Request(del); err != nil {
		log.Println("delete lesson:", err)
	}
}

func (h *Handler) sendOffer(chatID int64) {
	// Копируем продающее сообщение из канала и добавляем кнопку
	copy := tgbotapi.NewCopyMessage(chatID, salesChatID, salesMsgID)
	copy.ReplyMarkup = h.buyKeyboard()

	if _, err := h.bot.API.Request(copy); err != nil {
		log.Printf("copy sales message error: %v", err)
		// Просто логируем ошибку, не отправляем фоллбэк
	}
}

func (h *Handler) RunDeletionScheduler(ctx context.Context) {
	if h.store == nil {
		return
	}
	ticker := time.NewTicker(1 * time.Second) // быстрее для теста; в проде можно 5с
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()

			// 1) офферы
			offers, err := h.store.FetchDueOffers(ctx, now, 100)
			if err == nil {
				for _, t := range offers {
					// msgID не нужен для оффера — важен chatID
					h.sendOffer(t.ChatID)
				}
			}

			// 2) удаления
			dels, err := h.store.FetchDueDeletions(ctx, now, 100)
			if err == nil {
				for _, t := range dels {
					h.deleteLesson(t.ChatID, t.MsgID) // если хочешь «оффер после удаления» — оставь; или замени на просто delete
				}
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

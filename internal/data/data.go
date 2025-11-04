package data

type RefItem struct {
	FromChatID int64
	MessageID  int
}

// Теперь у ниши есть Gif (сообщение в канале), вместо Dir/локального файла.
type NicheDef struct {
	VisibleTitle string
	Emoji        string
	CaptionWord  string
	Gif          RefItem   // ← откуда копируем гифку
	Posts        []RefItem // «референсы»
}

// Порядок отображения ниш
var NicheOrder = []string{
	"Автомобили",
	"Недвижимость",
	"Кофейни/Кондитерские",
	"Услуги",
	"Бренды",
}

var Niches = map[string]NicheDef{
	"brands": {
		VisibleTitle: "Бренды",
		Emoji:        "🏷️",
		CaptionWord:  "бренды",
		Gif:          RefItem{FromChatID: -1003212181419, MessageID: 33}, // ← твой пост с гифкой
		Posts: []RefItem{
			{FromChatID: -1003212181419, MessageID: 25},
			{FromChatID: -1003212181419, MessageID: 19},
			{FromChatID: -1003212181419, MessageID: 16},
		},
	},
	"cafe": {
		VisibleTitle: "Кофейни/Кондитерские",
		Emoji:        "☕",
		CaptionWord:  "кофейни/кондитерские",
		Gif:          RefItem{FromChatID: -1003212181419, MessageID: 31},
		Posts: []RefItem{
			{FromChatID: -1003212181419, MessageID: 21},
			{FromChatID: -1003212181419, MessageID: 12},
			{FromChatID: -1003212181419, MessageID: 8},
		},
	},
	"cars": {
		VisibleTitle: "Автомобили",
		Emoji:        "🚗",
		CaptionWord:  "авто",
		Gif:          RefItem{FromChatID: -1003212181419, MessageID: 29},
		Posts: []RefItem{
			{FromChatID: -1003212181419, MessageID: 26},
			{FromChatID: -1003212181419, MessageID: 22},
			{FromChatID: -1003212181419, MessageID: 20},
		},
	},
	"immovables": {
		VisibleTitle: "Недвижимость",
		Emoji:        "🏠",
		CaptionWord:  "недвижимость",
		Gif:          RefItem{FromChatID: -1003212181419, MessageID: 30},
		Posts: []RefItem{
			{FromChatID: -1003212181419, MessageID: 24},
			{FromChatID: -1003212181419, MessageID: 15},
			{FromChatID: -1003212181419, MessageID: 10},
		},
	},
	"services": {
		VisibleTitle: "Услуги",
		Emoji:        "🧰",
		CaptionWord:  "услуги",
		Gif:          RefItem{FromChatID: -1003212181419, MessageID: 32},
		Posts: []RefItem{
			{FromChatID: -1003212181419, MessageID: 23},
			{FromChatID: -1003212181419, MessageID: 17},
			{FromChatID: -1003212181419, MessageID: 13},
		},
	},
}

// соответствие «видимое имя» → «ключ»
var NameToKey = map[string]string{
	"Автомобили":           "cars",
	"Недвижимость":         "immovables",
	"Кофейни/Кондитерские": "cafe",
	"Услуги":               "services",
	"Бренды":               "brands",
}

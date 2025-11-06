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
	"автомобили",
	"недвижимость",
	"кофейни/кондитерские",
	"услуги",
	"бренды",
}

var Niches = map[string]NicheDef{
	"brands": {
		VisibleTitle: "бренды",
		Emoji:        "🏷️",
		CaptionWord:  "бренды",
		Gif:          RefItem{FromChatID: -1003212181419, MessageID: 39}, // ← твой пост с гифкой
		Posts: []RefItem{
			{FromChatID: -1003212181419, MessageID: 25},
			{FromChatID: -1003212181419, MessageID: 19},
			{FromChatID: -1003212181419, MessageID: 16},
		},
	},
	"cafe": {
		VisibleTitle: "кофейни/кондитерские",
		Emoji:        "☕",
		CaptionWord:  "кофейни/кондитерские",
		Gif:          RefItem{FromChatID: -1003212181419, MessageID: 37},
		Posts: []RefItem{
			{FromChatID: -1003212181419, MessageID: 21},
			{FromChatID: -1003212181419, MessageID: 12},
			{FromChatID: -1003212181419, MessageID: 8},
		},
	},
	"cars": {
		VisibleTitle: "автомобили",
		Emoji:        "🚗",
		CaptionWord:  "авто",
		Gif:          RefItem{FromChatID: -1003212181419, MessageID: 35},
		Posts: []RefItem{
			{FromChatID: -1003212181419, MessageID: 26},
			{FromChatID: -1003212181419, MessageID: 22},
			{FromChatID: -1003212181419, MessageID: 20},
		},
	},
	"immovables": {
		VisibleTitle: "недвижимость",
		Emoji:        "🏠",
		CaptionWord:  "недвижимость",
		Gif:          RefItem{FromChatID: -1003212181419, MessageID: 36},
		Posts: []RefItem{
			{FromChatID: -1003212181419, MessageID: 24},
			{FromChatID: -1003212181419, MessageID: 15},
			{FromChatID: -1003212181419, MessageID: 10},
		},
	},
	"services": {
		VisibleTitle: "услуги",
		Emoji:        "🧰",
		CaptionWord:  "услуги",
		Gif:          RefItem{FromChatID: -1003212181419, MessageID: 38},
		Posts: []RefItem{
			{FromChatID: -1003212181419, MessageID: 23},
			{FromChatID: -1003212181419, MessageID: 17},
			{FromChatID: -1003212181419, MessageID: 13},
		},
	},
}

// соответствие «видимое имя» → «ключ»
var NameToKey = map[string]string{
	"автомобили":           "cars",
	"недвижимость":         "immovables",
	"кофейни/кондитерские": "cafe",
	"услуги":               "services",
	"бренды":               "brands",
}

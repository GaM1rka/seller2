package data

// Порядок отображения ниш в меню (видимые названия)
var NicheOrder = []string{
	"Автомобили",
	"Недвижимость",
	"Кофейни/Кондитерские",
	"Услуги",
	"Бренды",
}

type RefItem struct {
	FromChatID int64
	MessageID  int
}

// Ключи: brands, cafe, cars, immovables, services (совпадает с папками в /video)
var Niches = map[string]struct {
	VisibleTitle string // название в меню
	Emoji        string
	CaptionWord  string // первое слово в подписи ("недвижимость", "авто", ...)
	Dir          string // папка с гифкой
	Posts        []RefItem
}{
	"brands": {
		VisibleTitle: "Бренды",
		Emoji:        "🏷️",
		CaptionWord:  "бренды",
		Dir:          "video/brands",
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
		Dir:          "video/cafe",
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
		Dir:          "video/cars",
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
		Dir:          "video/immovables",
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
		Dir:          "video/services",
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

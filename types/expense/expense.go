package expense

const (
	OTHER = iota + 1
	CANCEL
	TELECOM
	FOOD
	CLOTHES
	CAR
	HEALTH
	CHEMICALS
	GIFTS
	COMMUNAL
	EDUCATION
	BEAUTY
	RELAX
)

var ExpenseMap = map[int]string{
	CANCEL:    "Відмінити",
	OTHER:     "Інше",
	TELECOM:   "Телеком",
	FOOD:      "Харчування",
	CLOTHES:   "Одяг",
	CAR:       "Авто",
	HEALTH:    "Здоров'я",
	CHEMICALS: "Хімія",
	GIFTS:     "Подарунки",
	COMMUNAL:  "Комуналка",
	EDUCATION: "Навчання",
	BEAUTY:    "Краса",
	RELAX:     "Відпочинок",
}

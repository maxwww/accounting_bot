package expense

const (
	OTHER = iota + 1
	TELECOM
	FOOD
	CLOTHES
	CAR
	HEALTH
	CHEMICALS
	GIFTS
	CANCEL
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
}

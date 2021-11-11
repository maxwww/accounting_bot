package expense

const (
	OTHER = iota + 1
	TELECOM
	FOOD
	CLOTHES
	CAR
	HEALTH
)

var ExpenseMap = map[int]string{
	OTHER:   "Інше",
	TELECOM: "Телеком",
	FOOD:    "Харчування",
	CLOTHES: "Одяг",
	CAR:     "Авто",
	HEALTH:  "Здоров'я",
}

package settings

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/maxwww/accounting_bot/types/expense"
)

func NewKeyboard() *tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Баланс"),
			tgbotapi.NewKeyboardButton("Поточні"),
			tgbotapi.NewKeyboardButton("Минулі"),
			tgbotapi.NewKeyboardButton("Позаминулі"),
		),
	)

	return &keyboard
}

func NewExpenseKeyboard(amount float64, now string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"Телеком",
				fmt.Sprintf("%d@@%.2f@@%s", expense.TELECOM, amount, now),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"Харчування",
				fmt.Sprintf("%d@@%.2f@@%s", expense.FOOD, amount, now),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"Одяг",
				fmt.Sprintf("%d@@%.2f@@%s", expense.CLOTHES, amount, now),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"Транспорт",
				fmt.Sprintf("%d@@%.2f@@%s", expense.CAR, amount, now),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"Здоров'я",
				fmt.Sprintf("%d@@%.2f@@%s", expense.HEALTH, amount, now),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"Відпочинок",
				fmt.Sprintf("%d@@%.2f@@%s", expense.RELAX, amount, now),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"Хімія",
				fmt.Sprintf("%d@@%.2f@@%s", expense.CHEMICALS, amount, now),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"Подарунки",
				fmt.Sprintf("%d@@%.2f@@%s", expense.GIFTS, amount, now),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"Комуналка",
				fmt.Sprintf("%d@@%.2f@@%s", expense.COMMUNAL, amount, now),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"Навчання",
				fmt.Sprintf("%d@@%.2f@@%s", expense.EDUCATION, amount, now),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"Краса",
				fmt.Sprintf("%d@@%.2f@@%s", expense.BEAUTY, amount, now),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"Інше",
				fmt.Sprintf("%d@@%.2f@@%s", expense.OTHER, amount, now),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"❌ Відмінити",
				fmt.Sprintf("%d@@%.2f@@%s", expense.CANCEL, amount, now),
			),
		),
	)
}

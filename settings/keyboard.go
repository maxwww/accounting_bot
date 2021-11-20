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
				"Авто",
				fmt.Sprintf("%d@@%.2f@@%s", expense.CAR, amount, now),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"Здоров'я",
				fmt.Sprintf("%d@@%.2f@@%s", expense.HEALTH, amount, now),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"Інше",
				fmt.Sprintf("%d@@%.2f@@%s", expense.OTHER, amount, now),
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
				"❌ Відмінити",
				fmt.Sprintf("%d@@%.2f@@%s", expense.CANCEL, amount, now),
			),
		),
	)
}

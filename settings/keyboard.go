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
			tgbotapi.NewKeyboardButton("Минуломісячні"),
		),
	)

	return &keyboard
}

func NewExpenseKeyboard(amount float64, now int64) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"Телеком",
				fmt.Sprintf("%d@@%.2f@@%d", expense.TELECOM, amount, now),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"Харчування",
				fmt.Sprintf("%d@@%.2f@@%d", expense.FOOD, amount, now),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"Одяг",
				fmt.Sprintf("%d@@%.2f@@%d", expense.CLOTHES, amount, now),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"Авто",
				fmt.Sprintf("%d@@%.2f@@%d", expense.CAR, amount, now),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"Здоров'я",
				fmt.Sprintf("%d@@%.2f@@%d", expense.HEALTH, amount, now),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"Інше",
				fmt.Sprintf("%d@@%.2f@@%d", expense.OTHER, amount, now),
			),
		),
	)
}

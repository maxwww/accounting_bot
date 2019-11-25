package settings

import "github.com/go-telegram-bot-api/telegram-bot-api"

func NewKeyboard() *tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Баланс"),
			tgbotapi.NewKeyboardButton("Налаштування"),
		),
	)

	return &keyboard
}

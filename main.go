package main

import (
	"database/sql"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"github.com/maxwww/accounting_bot/bank/privat"
	"github.com/maxwww/accounting_bot/db"
	"github.com/maxwww/accounting_bot/settings"
	"log"
	"os"
	"strconv"
)

var (
	dbConnection          *sql.DB
	token                 string
	merchant              string
	privatBalanceEndpoint string
	password              string
	privatCard            string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	token = os.Getenv("TOKEN")
	merchant = os.Getenv("MERCHANT")
	privatBalanceEndpoint = os.Getenv("PRIVAT_BALANCE_ENDPOINT")
	password = os.Getenv("PASSWORD")
	privatCard = os.Getenv("PRIVAT_CARD")

	dbConnection = db.NewConnection()
	defer func() {
		err := dbConnection.Close()
		if err != nil {
			log.Print(err)
		}
	}()
}

func main() {
	defer func() {
		err := dbConnection.Close()
		if err != nil {
			log.Print(err)
		}
	}()

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		go handleUpdate(update, bot)
	}
}

func handleUpdate(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if update.Message == nil {
		return
	}
	keyboard := settings.NewKeyboard()
	log.Printf("[%s] %v - start", update.Message.From.UserName, update.Message.Text)
	defer log.Printf("[%s] %s - end", update.Message.From.UserName, update.Message.Text)
	db.LogUser(dbConnection, update.Message.From.ID, update.Message.From.IsBot, update.Message.From.FirstName, update.Message.From.LastName, update.Message.From.UserName, update.Message.From.LanguageCode)

	switch update.Message.Text {
	case "Баланс":
		var responseMessage string
		balance, err := privat.GetBalance(password, privatCard, merchant, privatBalanceEndpoint)
		if err != nil {
			fmt.Println(err)
			responseMessage = "failed to handle request"
		} else {
			responseMessage = strconv.FormatFloat(balance, 'f', -1, 64)
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseMessage)
		msg.ReplyMarkup = keyboard
		_, err = bot.Send(msg)
		if err != nil {
			log.Print(err)
		}
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyMarkup = keyboard
		_, err := bot.Send(msg)
		if err != nil {
			log.Print(err)
		}
	}
}

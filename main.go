package main

import (
	"database/sql"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"github.com/maxwww/accounting_bot/bank/mono"
	"github.com/maxwww/accounting_bot/bank/privat"
	"github.com/maxwww/accounting_bot/db"
	"github.com/maxwww/accounting_bot/settings"
	"github.com/maxwww/accounting_bot/types"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	dbConnection          *sql.DB
	token                 string
	bot                   *tgbotapi.BotAPI
	keyboard              *tgbotapi.ReplyKeyboardMarkup
	merchantH             string
	merchantW             string
	passwordH             string
	passwordW             string
	privatBalanceEndpoint string
	privatCardH           string
	privatCardW           string
	monoInfoEndpoint      string
	monoTokenH            string
	monoTokenW            string
	idH                   int
	idW                   int
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	token = os.Getenv("TOKEN")
	merchantH = os.Getenv("MERCHANT_H")
	merchantW = os.Getenv("MERCHANT_W")
	privatBalanceEndpoint = os.Getenv("PRIVAT_BALANCE_ENDPOINT")
	passwordH = os.Getenv("PASSWORD_H")
	passwordW = os.Getenv("PASSWORD_W")
	privatCardH = os.Getenv("PRIVAT_CARD_H")
	privatCardW = os.Getenv("PRIVAT_CARD_W")
	monoInfoEndpoint = os.Getenv("MONO_API_ENDPOINT")
	monoTokenH = os.Getenv("MONO_H")
	monoTokenW = os.Getenv("MONO_W")
	idH, err = strconv.Atoi(os.Getenv("ID_H"))
	if err != nil {
		log.Fatal("Id is not specified")
	}
	idW, err = strconv.Atoi(os.Getenv("ID_W"))
	if err != nil {
		log.Fatal("Id is not specified")
	}

	dbConnection = db.NewConnection()

	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
	keyboard = settings.NewKeyboard()
}

func main() {
	defer func() {
		err := dbConnection.Close()
		if err != nil {
			log.Print(err)
		}
	}()

	c := cron.New(
		cron.WithLocation(time.UTC))
	c.AddFunc("0 * * * *", sendBalance)
	c.Start()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		go handleUpdate(update)
	}
}

func handleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}
	log.Printf("[%s] %v - start", update.Message.From.UserName, update.Message.Text)
	defer log.Printf("[%s] %s - end", update.Message.From.UserName, update.Message.Text)
	db.LogUser(dbConnection, update.Message.From.ID, update.Message.From.IsBot, update.Message.From.FirstName, update.Message.From.LastName, update.Message.From.UserName, update.Message.From.LanguageCode)

	if update.Message.From.ID != idH && update.Message.From.ID != idW {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyMarkup = keyboard
		_, err := bot.Send(msg)
		if err != nil {
			log.Print(err)
		}

		return
	}

	switch update.Message.Text {
	case "Баланс":
		var wg sync.WaitGroup
		balanceChan := make(chan *types.Balance)

		go func() {
			wg.Wait()
			close(balanceChan)
		}()

		wg.Add(4)
		go privat.GetBalance(passwordH, privatCardH, merchantH, privatBalanceEndpoint, balanceChan, &wg, "privat_h")
		go privat.GetBalance(passwordW, privatCardW, merchantW, privatBalanceEndpoint, balanceChan, &wg, "privat_w")
		go mono.GetBalance(monoTokenH, monoInfoEndpoint, balanceChan, &wg, "mono_h")
		go mono.GetBalance(monoTokenW, monoInfoEndpoint, balanceChan, &wg, "mono_w")

		responseMessage := ""
		balances := map[string]float64{
			"privat_h": 0,
			"mono_h":   0,
			"privat_w": 0,
			"mono_w":   0,
		}

		for balance := range balanceChan {
			if balance.Error != nil {
				responseMessage = "failed to handle request"
				fmt.Println(balance.Error)
				continue
			}
			balances[balance.Type] = balance.Balance
		}
		if responseMessage == "" {
			responseMessage = fmt.Sprintf(`Загальний баланс: _%.2f_
Максим Приват: _%.2f_
Максим Моно: _%.2f_
Оксана Приват: _%.2f_
Оксана Моно: _%.2f_
`,
				balances["privat_h"]+balances["mono_h"]+balances["privat_w"]+balances["mono_w"],
				balances["privat_h"],
				balances["mono_h"],
				balances["privat_w"],
				balances["mono_w"])
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseMessage)
		msg.ParseMode = "markdown"
		msg.DisableWebPagePreview = true
		msg.ReplyMarkup = keyboard
		_, err := bot.Send(msg)
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

// TODO: use one function
func sendBalance() {
	var wg sync.WaitGroup
	balanceChan := make(chan *types.Balance)

	go func() {
		wg.Wait()
		close(balanceChan)
	}()

	wg.Add(4)
	go privat.GetBalance(passwordH, privatCardH, merchantH, privatBalanceEndpoint, balanceChan, &wg, "privat_h")
	go privat.GetBalance(passwordW, privatCardW, merchantW, privatBalanceEndpoint, balanceChan, &wg, "privat_w")
	go mono.GetBalance(monoTokenH, monoInfoEndpoint, balanceChan, &wg, "mono_h")
	go mono.GetBalance(monoTokenW, monoInfoEndpoint, balanceChan, &wg, "mono_w")

	responseMessage := ""
	balances := map[string]float64{
		"privat_h": 0,
		"mono_h":   0,
		"privat_w": 0,
		"mono_w":   0,
	}

	for balance := range balanceChan {
		if balance.Error != nil {
			responseMessage = "failed to handle request"
			continue
		}
		balances[balance.Type] = balance.Balance
	}
	if responseMessage == "" {
		responseMessage = fmt.Sprintf(`Загальний баланс: _%.2f_
Максим Приват: _%.2f_
Максим Моно: _%.2f_
Оксана Приват: _%.2f_
Оксана Моно: _%.2f_
`,
			balances["privat_h"]+balances["mono_h"]+balances["privat_w"]+balances["mono_w"],
			balances["privat_h"],
			balances["mono_h"],
			balances["privat_w"],
			balances["mono_w"])
	}

	msg := tgbotapi.NewMessage(int64(idH), responseMessage)
	msg.ParseMode = "markdown"
	msg.DisableWebPagePreview = true
	msg.ReplyMarkup = keyboard
	_, err := bot.Send(msg)
	if err != nil {
		log.Print(err)
	}
}

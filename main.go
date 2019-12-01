package main

import (
	"database/sql"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"github.com/maxwww/accounting_bot/bank/mono"
	"github.com/maxwww/accounting_bot/bank/privat"
	"github.com/maxwww/accounting_bot/bank/ukrsib"
	"github.com/maxwww/accounting_bot/constants"
	"github.com/maxwww/accounting_bot/db"
	"github.com/maxwww/accounting_bot/settings"
	"github.com/maxwww/accounting_bot/types"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
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
	monoCurrencyEndpoint  string
	monoTokenH            string
	monoTokenW            string
	idH                   int
	idW                   int
	response              types.Response
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
	monoCurrencyEndpoint = os.Getenv("MONO_CURRENCY_ENDPOINT")
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
	c.AddFunc("0 7 * * *", makeSendBalance([]int{idH, idW}))
	c.AddFunc("0 * * * *", makeSendBalance([]int{idH}))
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

	switch {
	case update.Message.Text == "Баланс":
		now := int(time.Now().Unix())
		if response.ResponseMessage != "" && now-response.Time < constants.DELAY {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, response.ResponseMessage)
			msg.ParseMode = "markdown"
			msg.DisableWebPagePreview = true
			msg.ReplyMarkup = keyboard
			_, err := bot.Send(msg)
			if err != nil {
				log.Print(err)
			}
			break
		}

		var wg sync.WaitGroup
		balanceChan := make(chan *types.Balance)

		go func() {
			wg.Wait()
			close(balanceChan)
		}()

		wg.Add(5)
		go ukrsib.GetBalance(dbConnection, monoCurrencyEndpoint, balanceChan, &wg, 1)
		go privat.GetBalance(passwordH, privatCardH, merchantH, privatBalanceEndpoint, balanceChan, &wg, "Максим Приват", 4)
		go mono.GetBalance(monoTokenH, monoInfoEndpoint, balanceChan, &wg, "Максим Моно", 5)
		go privat.GetBalance(passwordW, privatCardW, merchantW, privatBalanceEndpoint, balanceChan, &wg, "Оксана Приват", 6)
		go mono.GetBalance(monoTokenW, monoInfoEndpoint, balanceChan, &wg, "Оксана Моно", 7)

		responseMessage := ""
		var balances []types.Balance

		totalBalance := .0
		for balance := range balanceChan {
			if balance.Error != nil {
				responseMessage = "Failed to get balance"

				totalBalance += balance.Balance
				balances = append(balances, *balance)
			} else {
				totalBalance += balance.Balance
				balances = append(balances, *balance)
			}
		}
		sort.Slice(balances, func(i, j int) bool { return balances[i].Order < balances[j].Order })

		if responseMessage == "" {
			responseFormat := "Загальний баланс: _%.2f_"
			responseParams := []interface{}{totalBalance}
			for _, v := range balances {
				if v.UsdBalance != 0 {
					responseFormat += "\n%s: _$%.2f_ (_%.2f_)"
					responseParams = append(responseParams, v.Name, v.UsdBalance, v.Balance)
				} else {
					responseFormat += "\n%s: _%.2f_"
					responseParams = append(responseParams, v.Name, v.Balance)
				}
			}

			responseMessage = fmt.Sprintf(responseFormat, responseParams...)
		}

		response.ResponseMessage = responseMessage
		response.Time = now

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseMessage)
		msg.ParseMode = "markdown"
		msg.DisableWebPagePreview = true
		msg.ReplyMarkup = keyboard
		_, err := bot.Send(msg)
		if err != nil {
			log.Print(err)
		}
	case strings.HasPrefix(update.Message.Text, "ukr"):
		responseMessage := "OK!"
		args := strings.Split(update.Message.Text, " ")
		if len(args) != 2 {
			responseMessage = "Parse error!"
		} else {
			f, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				responseMessage = "Parse error!"
			} else {
				err = db.UpdateAccount(dbConnection, args[0], f)
				if err != nil {
					responseMessage = "Parse error!"
				}
			}
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
func makeSendBalance(ids []int) func() {
	return func() {
		now := int(time.Now().Unix())
		if response.ResponseMessage != "" && now-response.Time < constants.DELAY {
			for _, id := range ids {
				msg := tgbotapi.NewMessage(int64(id), response.ResponseMessage)
				msg.ParseMode = "markdown"
				msg.DisableWebPagePreview = true
				msg.ReplyMarkup = keyboard
				_, err := bot.Send(msg)
				if err != nil {
					log.Print(err)
				}
			}
			return
		}

		var wg sync.WaitGroup
		balanceChan := make(chan *types.Balance)

		go func() {
			wg.Wait()
			close(balanceChan)
		}()

		wg.Add(5)
		go ukrsib.GetBalance(dbConnection, monoCurrencyEndpoint, balanceChan, &wg, 1)
		go privat.GetBalance(passwordH, privatCardH, merchantH, privatBalanceEndpoint, balanceChan, &wg, "Максим Приват", 4)
		go mono.GetBalance(monoTokenH, monoInfoEndpoint, balanceChan, &wg, "Максим Моно", 5)
		go privat.GetBalance(passwordW, privatCardW, merchantW, privatBalanceEndpoint, balanceChan, &wg, "Оксана Приват", 6)
		go mono.GetBalance(monoTokenW, monoInfoEndpoint, balanceChan, &wg, "Оксана Моно", 7)

		responseMessage := ""
		var balances []types.Balance

		totalBalance := .0
		for balance := range balanceChan {
			if balance.Error != nil {
				responseMessage = "Failed to get balance"

				totalBalance += balance.Balance
				balances = append(balances, *balance)
			} else {
				totalBalance += balance.Balance
				balances = append(balances, *balance)
			}
		}
		sort.Slice(balances, func(i, j int) bool { return balances[i].Order < balances[j].Order })

		if responseMessage == "" {
			responseFormat := "Загальний баланс: _%.2f_"
			responseParams := []interface{}{totalBalance}
			for _, v := range balances {
				if v.UsdBalance != 0 {
					responseFormat += "\n%s: _$%.2f_ (_%.2f_)"
					responseParams = append(responseParams, v.Name, v.UsdBalance, v.Balance)
				} else {
					responseFormat += "\n%s: _%.2f_"
					responseParams = append(responseParams, v.Name, v.Balance)
				}
			}

			responseMessage = fmt.Sprintf(responseFormat, responseParams...)
		}

		response.ResponseMessage = responseMessage
		response.Time = now

		for _, id := range ids {
			msg := tgbotapi.NewMessage(int64(id), responseMessage)
			msg.ParseMode = "markdown"
			msg.DisableWebPagePreview = true
			msg.ReplyMarkup = keyboard
			_, err := bot.Send(msg)
			if err != nil {
				log.Print(err)
			}
		}
	}
}

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
	exp "github.com/maxwww/accounting_bot/types/expense"
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
	sentMessages          map[string][][]int
	queueChanel           chan struct{}
	familyMap             map[int64]int64
	balancesMap           map[string]types.BalanceTracker
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

	sentMessages = make(map[string][][]int)
	queueChanel = make(chan struct{}, 1)
	familyMap = map[int64]int64{
		int64(idW): int64(idH),
		int64(idH): int64(idW),
	}
	balancesMap = map[string]types.BalanceTracker{}
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
	//c.AddFunc("1 * * * *", makeSendBalance([]int{idH}))
	c.AddFunc("@every 90s", getBalances)
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
	if update.Message == nil && update.CallbackQuery == nil {
		return
	}

	var user *tgbotapi.User
	var inputText string
	if update.Message != nil {
		user = update.Message.From
		inputText = update.Message.Text
	} else if update.CallbackQuery != nil {
		user = update.CallbackQuery.From
		inputText = update.CallbackQuery.Data
	}

	log.Printf("[%s] %v - start", user.UserName, inputText)
	defer log.Printf("[%s] %s - end", user.UserName, inputText)
	db.LogUser(dbConnection, user.ID, user.IsBot, user.FirstName, user.LastName, user.UserName, user.LanguageCode)

	if update.CallbackQuery != nil {
		parts := strings.Split(update.CallbackQuery.Data, "@@")
		expense, _ := strconv.Atoi(parts[0])
		stringExpense := exp.ExpenseMap[expense]
		amount, _ := strconv.ParseFloat(parts[1], 64)
		key := parts[2]
		keyParts := strings.Split(key, "_")
		timestamp, _ := strconv.Atoi(keyParts[1])

		if expense == exp.CANCEL {
			if _, ok := sentMessages[key]; ok {
				for _, v := range sentMessages[key] {
					if len(v) == 2 {
						msg := tgbotapi.NewDeleteMessage(int64(v[0]), v[1])
						_, err := bot.Send(msg)
						if err != nil {
							log.Print(err)
						}
					}

					msg2 := tgbotapi.NewMessage(int64(v[0]), "Охрана отмєна")
					_, err := bot.Send(msg2)
					if err != nil {
						log.Print(err)
					}
				}

				delete(sentMessages, key)
			} else {
				msg := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
				_, err := bot.Send(msg)
				if err != nil {
					log.Print(err)
				}
			}

			return
		}

		location, _ := time.LoadLocation("Europe/Kiev")
		tm := time.Unix(int64(timestamp), 0).In(location)
		added := db.AddExpense(dbConnection, stringExpense, amount, tm, update.CallbackQuery.From.ID)

		if added {
			if _, ok := sentMessages[key]; ok {
				for _, v := range sentMessages[key] {
					if len(v) == 2 {
						msg := tgbotapi.NewDeleteMessage(int64(v[0]), v[1])
						_, err := bot.Send(msg)
						if err != nil {
							log.Print(err)
						}
					}

					msg2 := tgbotapi.NewMessage(int64(v[0]), fmt.Sprintf("_%.2f грн_ було витрачено на _%s_ (%s)", amount, stringExpense, update.CallbackQuery.From.FirstName))
					msg2.ParseMode = "markdown"
					msg2.DisableWebPagePreview = true
					_, err := bot.Send(msg2)
					if err != nil {
						log.Print(err)
					}
				}

				delete(sentMessages, key)
			} else {
				msg := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
				_, err := bot.Send(msg)
				if err != nil {
					log.Print(err)
				}
			}
		} else {
			delete(sentMessages, key)
		}

		return
	}

	if update.Message.From.ID != idH && update.Message.From.ID != idW {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyMarkup = keyboard
		_, err := bot.Send(msg)
		if err != nil {
			log.Print(err)
		}

		return
	}

	preparedText := strings.Replace(update.Message.Text, ",", ".", -1)
	parsedFloat, floatErr := strconv.ParseFloat(preparedText, 64)
	switch {
	case update.Message.Text == "Баланс":
		now := int(time.Now().Unix())
		getBalances()
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
	case update.Message.Text == "Поточні":
		expenses, err := db.GetCurrentExpenses(dbConnection)
		if err != nil {
			log.Print(err)
		} else {
			if len(expenses) == 0 {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Витрати відсутні")
				msg.ReplyMarkup = keyboard
				_, err := bot.Send(msg)
				if err != nil {
					log.Print(err)
				}
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, buildExpenseMessage(expenses, "Ваші поточні витрати"))
				msg.ParseMode = "markdown"
				msg.ReplyMarkup = keyboard
				_, err := bot.Send(msg)
				if err != nil {
					log.Print(err)
				}
			}
		}
	case update.Message.Text == "Минулі":
		expenses, err := db.GetLastMonthExpenses(dbConnection)
		if err != nil {
			log.Print(err)
		} else {
			if len(expenses) == 0 {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Витрати відсутні")
				msg.ReplyMarkup = keyboard
				_, err := bot.Send(msg)
				if err != nil {
					log.Print(err)
				}
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, buildExpenseMessage(expenses, "Ваші минуломісячні витрати"))
				msg.ParseMode = "markdown"
				msg.ReplyMarkup = keyboard
				_, err := bot.Send(msg)
				if err != nil {
					log.Print(err)
				}
			}
		}
	case floatErr == nil:
		if parsedFloat > 0 {
			now := time.Now().Unix()
			key := fmt.Sprintf("%d_%d", update.Message.MessageID, now)
			stringBalance := fmt.Sprintf("%.2f", parsedFloat)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ваш баланс зменшився на _%s грн_", stringBalance))
			msg.ParseMode = "markdown"
			msg.DisableWebPagePreview = true
			expenseKeyboard := settings.NewExpenseKeyboard(parsedFloat, key)
			msg.ReplyMarkup = &expenseKeyboard
			sent, err := bot.Send(msg)
			if err != nil {
				log.Print(err)
			}
			sentMessages[key] = append(sentMessages[key], []int{int(sent.Chat.ID), sent.MessageID})
			if partnerId, ok := familyMap[sent.Chat.ID]; ok {
				sentMessages[key] = append(sentMessages[key], []int{int(partnerId)})
			}
		}
	default:
		msgD := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID)
		_, err := bot.Send(msgD)
		if err != nil {
			log.Print(err)
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyMarkup = keyboard
		_, err = bot.Send(msg)
		if err != nil {
			log.Print(err)
		}

		if partnerId, ok := familyMap[update.Message.Chat.ID]; ok {
			msg := tgbotapi.NewMessage(partnerId, update.Message.Text+fmt.Sprintf(" (від %s)", update.Message.From.FirstName))
			msg.ReplyMarkup = keyboard
			_, err := bot.Send(msg)
			if err != nil {
				log.Print(err)
			}
		}
	}
}

func sendExpense(balanceDiff float64, prefix string) {
	ids := []int{idH, idW}
	//ids := []int{idH}
	now := time.Now().Unix()
	key := fmt.Sprintf("%s_%d", prefix, now)
	stringBalance := fmt.Sprintf("%.2f", balanceDiff)
	if balanceDiff > 0 && stringBalance != "0.00" {
		sentMessages[key] = [][]int{}

		for _, id := range ids {
			msg := tgbotapi.NewMessage(int64(id), fmt.Sprintf("Ваш баланс зменшився на _%s грн_", stringBalance))
			msg.ParseMode = "markdown"
			msg.DisableWebPagePreview = true
			expenseKeyboard := settings.NewExpenseKeyboard(balanceDiff, key)
			msg.ReplyMarkup = &expenseKeyboard
			sent, err := bot.Send(msg)
			if err != nil {
				log.Print(err)
			}
			sentMessages[key] = append(sentMessages[key], []int{int(sent.Chat.ID), sent.MessageID})
		}
	}
}

func makeSendBalance(ids []int) func() {
	return func() {
		now := int(time.Now().Unix())
		getBalances()
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
	}
}

func getBalances() {
	queueChanel <- struct{}{}
	now := int(time.Now().Unix())
	if response.ResponseMessage == "" || now-response.Time > constants.DELAY {
		var wg sync.WaitGroup
		balanceChan := make(chan *types.Balance)
		wg.Add(5)
		go func() {
			wg.Wait()
			close(balanceChan)
		}()

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
				fmt.Println(balance.Error)
				responseMessage = "Failed to get balance"
			} else {
				if balance.CheckExpense {
					if fmt.Sprintf("%.2f", balance.Balance) == "0.00" {
						fmt.Println("empty balance")
						fmt.Println(balance)
					} else {
						if _, ok := balancesMap[balance.Name]; !ok {
							balancesMap[balance.Name] = types.BalanceTracker{
								Amount:     balance.Balance,
								PrevAmount: 0,
							}
						}

						balancesMap[balance.Name] = types.BalanceTracker{
							Amount:     balance.Balance,
							PrevAmount: balancesMap[balance.Name].Amount,
						}

						if balancesMap[balance.Name].PrevAmount != 0 && balancesMap[balance.Name].Amount != 0 && balancesMap[balance.Name].PrevAmount != balancesMap[balance.Name].Amount {
							sendExpense(balancesMap[balance.Name].PrevAmount-balancesMap[balance.Name].Amount, balance.Name)
						}
					}
				}
			}
			totalBalance += balance.Balance
			balances = append(balances, *balance)
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
	}
	<-queueChanel
}

func buildExpenseMessage(expenses []types.Expense, startMessage string) string {
	message := startMessage + ":\n"
	total := .0
	for _, v := range expenses {
		message += fmt.Sprintf("\n*%s* - _%.2f грн_", v.Expense, v.Amount)
		total += v.Amount
	}
	message += fmt.Sprintf("\n--------------------\nВсього: _%.2f грн_", total)

	return message
}

package mono

import (
	"encoding/json"
	"github.com/maxwww/accounting_bot/types"
	"io/ioutil"
	"net/http"
	"sync"
)

type Response struct {
	Name     string    `json:"name"`
	Accounts []Account `json:"accounts"`
}

type Account struct {
	Balance     int    `json:"balance"`
	CreditLimit int    `json:"creditLimit"`
	Type        string `json:"type"`
}

func GetBalance(token string, url string, ch chan *types.Balance, wg *sync.WaitGroup, name string, order int) {
	defer wg.Done()

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		ch <- &types.Balance{Error: err}
		return
	}
	req.Header.Add("X-Token", token)
	resp, err := client.Do(req)
	if err != nil {
		ch <- &types.Balance{Error: err}
		return
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ch <- &types.Balance{Error: err}
		return
	}

	response := Response{}
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		ch <- &types.Balance{Error: err}
		return
	}

	var balance float64
	var creditLimit float64

	for _, v := range response.Accounts {
		if v.Type == "black" {
			balance += float64(v.Balance)
			creditLimit += float64(v.CreditLimit)
		}
	}

	result := (balance - creditLimit) / 100

	ch <- &types.Balance{Balance: result, Name: name, Order: order, CheckExpense: true}
}

func GetRates(monoCurrencyEndpoint string, wg *sync.WaitGroup, ch chan *types.Rates) {
	defer wg.Done()

	res, err := http.Get(monoCurrencyEndpoint)
	if err != nil {
		ch <- &types.Rates{Error: err}
		return
	}

	response, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		ch <- &types.Rates{Error: err}
		return
	}

	var rates []types.Rate
	err = json.Unmarshal(response, &rates)
	if err != nil {
		ch <- &types.Rates{Error: err}
		return
	}

	ch <- &types.Rates{Rates: rates}
}

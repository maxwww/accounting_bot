package ukrsib

import (
	"database/sql"
	"github.com/maxwww/accounting_bot/bank/mono"
	"github.com/maxwww/accounting_bot/constants"
	"github.com/maxwww/accounting_bot/types"
	"sync"
)

func GetBalance(db *sql.DB, monoCurrencyEndpoint string, ch chan *types.Balance, wg *sync.WaitGroup, order int) {
	defer wg.Done()

	var (
		slug     string
		name     string
		currency int
		balance  float64
		wgRates  sync.WaitGroup
	)

	ratesChanel := make(chan *types.Rates)
	wgRates.Add(1)
	go func() {
		wgRates.Wait()
		close(ratesChanel)
	}()
	go mono.GetRates(monoCurrencyEndpoint, &wgRates, ratesChanel)

	rates := map[int]map[int]float64{}

	for v := range ratesChanel {
		if v.Error != nil {
			ch <- &types.Balance{Error: v.Error}
			return
		}

		for _, rate := range v.Rates {
			if _, ok := rates[rate.CurrencyCodeA]; !ok {
				rates[rate.CurrencyCodeA] = map[int]float64{}
			}
			rates[rate.CurrencyCodeA][rate.CurrencyCodeB] = rate.RateBuy
		}
	}

	rows, err := db.Query("select slug, name, currency, balance from accounts order by priority")
	if err != nil {
		ch <- &types.Balance{Error: err}
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&slug, &name, &currency, &balance)
		if err != nil {
			ch <- &types.Balance{Error: err}
			return
		}

		result := types.Balance{Balance: balance, Name: name, Order: order}
		order++
		if currency == constants.USD_CURRENCY {
			result.Balance = balance * rates[currency][980]
			result.UsdBalance = balance
		}
		ch <- &result
	}
	err = rows.Err()
	if err != nil {
		ch <- &types.Balance{Error: err}
		return
	}
}

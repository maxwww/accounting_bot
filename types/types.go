package types

type Balance struct {
	Balance    float64
	UsdBalance float64
	Name       string
	Error      error
	Order      int
}

type Rate struct {
	CurrencyCodeA int     `json:"currencyCodeA"`
	CurrencyCodeB int     `json:"currencyCodeB"`
	Date          int     `json:"date"`
	RateBuy       float64 `json:"rateBuy"`
	RateSell      float64 `json:"rateSell"`
}

type Rates struct {
	Rates []Rate
	Error error
}

type Response struct {
	Time            int
	ResponseMessage string
}

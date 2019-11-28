package privat

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"github.com/maxwww/accounting_bot/types"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Result struct {
	XMLName    xml.Name   `xml:"response"`
	ResultData ResultData `xml:"data"`
}

type ResultData struct {
	XMLName xml.Name `xml:"data"`
	Info    Info     `xml:"info"`
}

type Info struct {
	XMLName     xml.Name    `xml:"info"`
	Cardbalance Cardbalance `xml:cardbalance`
}

type Cardbalance struct {
	XMLName xml.Name `xml:"cardbalance"`
	Balance float64  `xml:"balance"`
}

const dataTemplate = `<oper>cmt</oper>
        <wait>20</wait>
        <test>0</test>
        <payment id="%d">
            <prop name="cardnum" value="%s" />
            <prop name="country" value="UA" />
        </payment>`
const xmlTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<request version="1.0">
    <merchant>
        <id>%s</id>
        <signature>%s</signature>
    </merchant>
    <data>%s</data>
</request>`

func GetBalance(password string, card string, merchant string, balanceUrl string, ch chan *types.Balance, wg *sync.WaitGroup, key string) {
	defer wg.Done()
	data := fmt.Sprintf(dataTemplate, time.Now().Second(), card)
	md5H := md5.New()
	sha1H := sha1.New()
	io.WriteString(md5H, fmt.Sprintf("%s%s", data, password))
	io.WriteString(sha1H, fmt.Sprintf("%x", md5H.Sum(nil)))
	signature := fmt.Sprintf("%x", sha1H.Sum(nil))
	xmlBody := fmt.Sprintf(xmlTemplate, merchant, signature, data)
	resp, err := http.Post(balanceUrl, "application/xml", strings.NewReader(xmlBody))
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
	result := Result{}
	err = xml.Unmarshal(bodyBytes, &result)
	if err != nil {
		ch <- &types.Balance{Error: err}
		return
	}
	fmt.Println(string(bodyBytes))
	ch <- &types.Balance{Balance: result.ResultData.Info.Cardbalance.Balance, Type: key}
}

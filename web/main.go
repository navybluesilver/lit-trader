package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"strconv"
	"time"
	config "github.com/navybluesilver/lit-trader/config"
	trader "github.com/navybluesilver/lit-trader/trader"
	counterparty "github.com/navybluesilver/lit-trader/counterparty"

)

const (
	coinType    uint32 = 1
)


var (
	templates        = template.Must(template.ParseFiles("template/orderbook.html"))
	certFile         = config.GetString("web.certFile")
	keyFile          = config.GetString("web.keyFile")
	port             = config.GetString("web.port")
	tName  			     = config.GetString("trader.name")
	tHost  			     = config.GetString("trader.ip")
	tPort  			     = config.GetInt("trader.port")
	mLNAddress       = config.GetString("counterparty.LNAddress")
	mName       		 = config.GetString("counterparty.name")
	mURL      		   = config.GetString("counterparty.url")
	mHost            = config.GetString("counterparty.ip")
	mPort            = config.GetInt("counterparty.port")

	fmap             = template.FuncMap{
		"formatAsSatoshi": formatAsSatoshi,
	}

)
var c *counterparty.Counterparty
var t *trader.Trader

type OrderbookPage struct {
	TraderName string
	TraderFullAddress string
	CounterpartyName string
	CounterpartyFullAddress string
	Instrument string
	Underlying string
	SettlementDate string
	SPOT int
	OracleName string
	OraclePubKey string
	OracleRpoint string
	OracleURL string
	Cash int
	PublicKey string
	WitnessPublicKey string
	Positions interface{}
	Offers interface{}
	Bids interface{}
	Asks interface{}
}

func main() {
	// intialise trader, counterparty and connect them
	c = &counterparty.Counterparty{Name: mName, LNAddress: mLNAddress, IP: mHost, Port: mPort, URL: mURL }
	t = trader.NewTrader(tName, tHost, tPort)
	//connect(t)

	//handling
	//orderbook
	http.HandleFunc("/", orderbookHandler)
	http.HandleFunc("/buy", buyHandler)
	http.HandleFunc("/sell", sellHandler)

	//files
	http.Handle("/template/", http.StripPrefix("/template/", http.FileServer(http.Dir("template"))))

	//listen

	// redirect every http request to https
	go http.ListenAndServe(port, http.HandlerFunc(redirect))
	log.Fatal(http.ListenAndServe(port, nil))
}

// Connect to the counterparty
func connect(t *trader.Trader) {
	fmt.Printf("[%s]- Connecting %s to %s [%s@%s:%d]\n", time.Now().Format("20060102150405"), t.Name, c.Name, c.LNAddress,  c.IP,  c.Port)
	err := t.Lit.Connect( c.LNAddress,  c.IP,  uint32(c.Port))
	handleError(err)
}

func redirect(w http.ResponseWriter, req *http.Request) {
	// remove/add not default ports from req.Host
	host := strings.Split(req.Host, ":")[0]
	target := "https://" + host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	log.Printf("redirect to: %s", target)
	http.Redirect(w, req, target,
	http.StatusTemporaryRedirect)
}

//orderbook
func orderbookHandler(w http.ResponseWriter, r *http.Request) {
	var o OrderbookPage
	o.TraderName = t.Name
	o.TraderFullAddress = t.TraderFullAddress()
	o.CounterpartyFullAddress = fmt.Sprintf("%s@%s:%d", mLNAddress, mHost, mPort)
	o.CounterpartyName = c.Name
	o.PublicKey = t.GetLegacyAddress()
	o.WitnessPublicKey = t.GetWitnessAddress()
	o.Instrument = t.GetInstrument()
	o.Underlying = t.GetUnderlying()
	o.OracleName = t.GetOracleName()
	o.OracleURL = t.GetOracleURL()
	o.OraclePubKey = t.GetOraclePubKey()
	o.OracleRpoint = trader.GetR(trader.GetSettlementTime())
	o.Bids = t.GetBids()
	o.Asks = t.GetAsks()
	o.Offers, _ = t.GetOffers()
	o.SPOT = t.GetCurrentSpot()
	o.SettlementDate = fmt.Sprintf("%v",time.Unix(int64(trader.GetSettlementTime()), 0))
	o.Cash = t.GetBalance(coinType)

	t := template.Must(template.New("orderbook.html").Funcs(fmap).ParseFiles("template/orderbook.html"))
	err := t.ExecuteTemplate(w, "orderbook.html", o)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func sellHandler(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		price, _ := strconv.Atoi(fmt.Sprintf("%s", r.Form["price"][0]))
		quantity, _ := strconv.Atoi(fmt.Sprintf("%s", r.Form["quantity"][0]))
		t.Sell(price, quantity)
		orderbookHandler(w, r)
}

func buyHandler(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		price, _ := strconv.Atoi(fmt.Sprintf("%s", r.Form["price"][0]))
		quantity, _ := strconv.Atoi(fmt.Sprintf("%s", r.Form["quantity"][0]))
		t.Buy(price, quantity)
		orderbookHandler(w, r)
}

//formatting
func formatAsSatoshi(satoshi float64) (string, error) {
	if satoshi == 0 {
		return "", nil
	}
	return fmt.Sprintf("%.0f", satoshi), nil
}

//error handling
func handleError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

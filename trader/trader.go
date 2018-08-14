package trader

import (
	"fmt"
	"github.com/mit-dci/lit-rpc-client-go"
	orderbook "github.com/navybluesilver/lit-trader/orderbook"
	config "github.com/navybluesilver/lit-trader/config"
	"time"
)

var (
	coinType       int    = config.GetInt("trader.coin_type")
)

type Trader struct {
	Name string
	Host string
	Port int
	Lit  *litrpcclient.LitRpcClient
}

// Return a new trader
// Import the oracle if it does not yet exist
func NewTrader(name string, host string, port int) (*Trader) {
	t := new(Trader)
	t.Name = name
	t.Host = host
	t.Port = port
	l, err := litrpcclient.NewClient(host, int32(port))
	handleError(err)
	t.Lit = l

	exists, err := t.oracleExists()
	if exists != true {
		t.Lit.ImportOracle(oracleUrl, oracleName)
	}
	return t
}

// Get the balance for a coinType and an address to allow deposits
func (t *Trader) GetBalance(coinType uint32) (int) {
	addr, err := t.Lit.GetAddresses(coinType, 0, false)
	handleError(err)

	allBal, err := t.Lit.ListBalances()
	handleError(err)

	for _, b := range allBal {
		if b.CoinType == coinType {
			fmt.Printf("[%s]- Trader: %s | CoinType: %d | SyncHeight: %d | Utxos: %d | WitConf: %d| Channel: %d | Address: %s\n", time.Now().Format("20060102150405"), t.Name, b.CoinType, b.SyncHeight, b.TxoTotal, b.MatureWitty, b.ChanTotal, addr)
			return int(b.MatureWitty)
		}
	}
	return 0
}

// Accept profitable trades
func (m *Trader) MakeMarket(port uint32) error {
	fmt.Printf("[%s]- Running: %v\n", time.Now().Format("20060102150405"), m.Name)

	// Listen for new offers
	err := m.Lit.Listen(fmt.Sprintf(":%d", port))
	handleError(err)

	// Endless loop to check if there are any contracts to accept
	for loop := true; loop; loop = (1 == 1) {
		allOrders, err := m.getAsksBids()
		handleError(err)
		c, err := orderbook.GetContractsToAccept(allOrders)
		if err != nil {
			if err.Error() == "Nothing to accept." {
				fmt.Printf("[%s]- %v\n", time.Now().Format("20060102150405"), err.Error())
				time.Sleep(10000 * time.Millisecond)
			} else {
				handleError(err)
			}
		} else {
			// TODO: handle not enough satoshis to accept contract
			for _, i := range c {
				err := m.Lit.AcceptContract(uint64(i))
				handleError(err)
				fmt.Printf("[%s]- Accepted contract [%v]\n", time.Now().Format("20060102150405"), i)
			}
		}

		m.SettleExpired()
	}
	return nil
}


// Settle all the contracts that are past the settlement date
func (t *Trader) SettleExpired() {
	//Get all Contracts
	allContracts, err := t.Lit.ListContracts()
	handleError(err)

	for _, c := range allContracts {
		// contracts active, past settlement date
		if c.Status == 5 && int(c.OracleTimestamp) < int(time.Now().Unix()) {
			v, s := GetOracleSignature()
			t.Lit.SettleContract(c.Idx, v, s)
			fmt.Printf("Settle contract [%v] at %v satoshis\n", c.Idx, v)
		}
	}
}

func (t *Trader) GetInstrument() (string) {
	return config.GetString("instrument.name")
}

func (t *Trader) GetUnderlying() (string) {
	return config.GetString("instrument.underlying")
}

func GetSettlementTime() (int) {
	return config.GetInt("instrument.settlement_time") // 17500 a76c8b4f6fe5770afffb0ad51adf0702d3b666ceee118eca6fbf27e0da9e6024
}

func GetMargin() (int) {
	return	config.GetInt("instrument.margin")
}

func (t *Trader) TraderFullAddress() (string) {
	addr, err := t.Lit.GetLNAddress()
	handleError(err)
	return fmt.Sprintf("%s@%s:%d", addr, t.Host, t.Port)
}

func (t *Trader) GetLegacyAddress() (pubKey string) {
  addr, err := t.Lit.GetAddresses(uint32(coinType), 0, true)
  handleError(err)
  return addr[0]
}

func (t *Trader) GetWitnessAddress() (pubKey string) {
  addr, err := t.Lit.GetAddresses(uint32(coinType), 0, false)
  handleError(err)
  return addr[0]
}

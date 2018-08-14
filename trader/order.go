package trader

import (
	"fmt"
	"time"
	orderbook "github.com/navybluesilver/lit-trader/orderbook"
)

// Buy future for the given price and quantity
// Calculate the funding and division based of the configured margin
func (t *Trader) Buy(price, quantity int) error {
	margin :=  GetMargin()
	fmt.Printf("[%s]- %s offers to buy %d items at %d xBT\n", time.Now().Format("20060102150405"), t.Name, quantity, price)
	ourFunding := int64(price * quantity)
	theirFunding := int64((price * quantity) * margin)
	valueFullyOurs := int64(0)
	valueFullyTheirs := int64(price + (price * margin))
	t.sendContract(ourFunding, theirFunding, valueFullyOurs, valueFullyTheirs)
	return nil
}

// Sell future for the given price and quantity
// Calculate the funding and division based of the configured margin
func (t *Trader) Sell(price, quantity int) error {
	margin :=  GetMargin()
	fmt.Printf("[%s]- %s offers to sell %d items at %d xBT\n", time.Now().Format("20060102150405"), t.Name, quantity, price)
	ourFunding := int64((price * quantity) * margin)
	theirFunding := int64(price * quantity)
	valueFullyOurs := int64(price + (price * margin))
	valueFullyTheirs := int64(0)
	t.sendContract(ourFunding, theirFunding, valueFullyOurs, valueFullyTheirs)
	return nil
}


// Return Bids
func (m *Trader) GetBids() (bids []orderbook.Order) {
	allOrders, err := m.getAsksBids()
	handleError(err)
	bids, _, _, _, err = orderbook.GetBidsAsks(allOrders)
	handleError(err)
	return bids
}

// Return Asks
func (m *Trader) GetAsks() (asks []orderbook.Order) {
	allOrders, err := m.getAsksBids()
	handleError(err)
	_, asks, _, _, err = orderbook.GetBidsAsks(allOrders)
	handleError(err)
	return asks
}

// Return the market maker peer index
// TODO: should not be hardcoded to 1
func (t *Trader) getMarketMakerIdx() (uint32, error) {
	return 1, nil
}

// Converts a lit contract into an ask or bid order
func (t *Trader) convertContractToOrder(contractIdx uint64) (o orderbook.Order, err error) {
	margin :=  GetMargin()
	// Get the contract from lit and copy the peerIdx and ContractIdx to the order
	c, err := t.Lit.GetContract(contractIdx)
	handleError(err)
	o.PeerIdx = int(c.PeerIdx)
	o.ContractIdx = int(c.Idx)
	o.Status = int(c.Status)


	// Make sure that both parties provide funding
	if c.OurFundingAmount == 0 && c.TheirFundingAmount == 0 {
		return o, fmt.Errorf("OurFundingAmount and TheirFundingAmount cannot both be 0")
	}

	// identify if it is a bid or an ask based on the funding
	if c.OurFundingAmount < c.TheirFundingAmount {
		o.AskBidInd = "ASK" // asking a certain price for the instrument, usually higher than the market price, usually triggered by a sell order
	} else {
		o.AskBidInd = "BID" // bidding a certain price for the instrument, usually lower than the market price, usually triggered by a buy order
	}

	// identify valueFullyOurs and valueFullyTheirs
	var valueFullyOurs int64
	var valueFullyTheirs int64

	// valueFullyTheirs is the minimum oracle value that gives us 0
	for _, d := range c.Division {
		if d.ValueOurs == 0 {
			if valueFullyTheirs == 0 {
				valueFullyTheirs = d.OracleValue
			}
			if d.OracleValue <= valueFullyTheirs {
				valueFullyTheirs = d.OracleValue
			}
		}
	}

	// valueFullyOurs is the minimum oracle value that gives us OurFundingAmount + TheirFundingAmount
	for _, d := range c.Division {
		if d.ValueOurs == c.OurFundingAmount+c.TheirFundingAmount {
			if valueFullyOurs == 0 {
				valueFullyOurs = d.OracleValue
			}
			if d.OracleValue <= valueFullyOurs {
				valueFullyOurs = d.OracleValue
			}
		}
	}

	if o.AskBidInd == "ASK" {
		// if it is a ask,
		// valueFullyOurs should be 0
		// valueFullyTheirs should not be 0
		if valueFullyOurs != 0 {
			return o, fmt.Errorf("valueFullyOurs for a ask should be 0")
		}
		if valueFullyTheirs == 0 {
			return o, fmt.Errorf("valueFullyTheirs for a ask should not be 0")
		}
		o.Price = int(valueFullyTheirs) / (1 + margin)
		o.Quantity = int(c.OurFundingAmount) / o.Price
	} else {
		// if its is an bid
		// valueFullyOurs should not be 0
		// valueFullyTheirs should be 0
		if valueFullyOurs == 0 {
			return o, fmt.Errorf("valueFullyOurs for a bid should not be 0")
		}

		if valueFullyTheirs != 0 {
			return o, fmt.Errorf("valueFullyTheirs for an bid should be 0")
		}
		o.Price = int(valueFullyOurs) / (1 + margin)
		o.Quantity = int(c.TheirFundingAmount) / o.Price
	}

	return o, nil
}

/*
//Order
ContractStatusOfferedByMe  DlcContractStatus = 1

//Ask/Bid
ContractStatusOfferedToMe  DlcContractStatus = 2

//Position
ContractStatusActive       DlcContractStatus = 6

//Ignored
ContractStatusDraft        DlcContractStatus = 0
ContractStatusDeclined     DlcContractStatus = 3
ContractStatusAccepted     DlcContractStatus = 4
ContractStatusAcknowledged DlcContractStatus = 5
ContractStatusSettling     DlcContractStatus = 7
ContractStatusClosed       DlcContractStatus = 8
ContractStatusError        DlcContractStatus = 9
ContractStatusAccepting    DlcContractStatus = 10
*/


//Return all open orders; get all contracts offerd by me
func (t *Trader) GetOffers() (orders []orderbook.Order, err error) {
	return t.getOrdersForContracts(1)
}

//Return all positions; get all active contracts
func (t *Trader) GetPositions() (orders []orderbook.Order, err error) {
	return t.getOrdersForContracts(6)
}

//Return all asks and bids; get all contracts offered to us
func (t *Trader) getAsksBids() (orders []orderbook.Order, err error) {
	return t.getOrdersForContracts(2)
}


//Return all buy and sell offers
func (t *Trader) getOrdersForContracts(status int) (orders []orderbook.Order, err error) {
	//Get all Contracts
	allContracts, err := t.Lit.ListContracts()
	handleError(err)

	//Loop all contracts offered to t
	//TODO: should not have to loop through all contracts one by oracle
	//TODO: should have the ability to completly remove contracts
	for _, c := range allContracts {
		// contracts offered to me
		if int(c.Status) == status {
			o, err := t.convertContractToOrder(c.Idx)
			if err != nil {
				//TODO: If the contract is not a valid order, decline it
				t.Lit.DeclineContract(c.Idx)
				fmt.Printf("Declined contract [%v]: %v\n", c.Idx, err)
			} else {
				orders = append(orders, o)
			}
		}
	}
	return orders, nil
}


// Create and offer the contract to the market maker
func (t *Trader) sendContract(ourFunding, theirFunding, valueFullyOurs, valueFullyTheirs int64) error {
	// Create a new empty draft contract
	contract, err := t.Lit.NewContract()
	handleError(err)

	// Get oracle
	oracleIdx, err := t.getOracleIdx(oracleName)
	handleError(err)

	// Configure the contract to use the oracle we need
	err = t.Lit.SetContractOracle(contract.Idx, oracleIdx)
	handleError(err)

	// Set the settlement time
	err = t.Lit.SetContractSettlementTime(contract.Idx, uint64(GetSettlementTime()))
	handleError(err)

	// Set the coin type of the contract
	err = t.Lit.SetContractCoinType(contract.Idx, uint32(coinType))
	handleError(err)

	// Configure the contract datafeed
	err = t.Lit.SetContractDatafeed(contract.Idx, uint64(datasourceId))
	handleError(err)

	// Set the contract funding to 1 BTC each
	err = t.Lit.SetContractFunding(contract.Idx, ourFunding, theirFunding)
	handleError(err)

	// Configure the contract division so that Alice get all the
	// funds when the value is 45000, and Bob gets
	// all the funds when the value is 1
	err = t.Lit.SetContractDivision(contract.Idx, valueFullyOurs, valueFullyTheirs)
	handleError(err)

	// Offer the contract to the market maker
	peerIdx, err := t.getMarketMakerIdx()
	err = t.Lit.OfferContract(contract.Idx, peerIdx)
	handleError(err)

	fmt.Printf("[%s]- %s offers contract: ourFunding [%d] | theirFunding [%d] | valueFullyOurs [%d] | valueFullyTheirs [%d]\n", time.Now().Format("20060102150405"), t.Name, ourFunding, theirFunding, valueFullyOurs, valueFullyTheirs)

	return nil
}

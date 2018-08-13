package orderbook

import (
	"fmt"
	"sort"
)

type Order struct {
	PeerIdx     int
	ContractIdx int
	AskBidInd   string
	Price       int
	Quantity    int
	Status			int
}

type Aggregates []struct {
	AggregateIdx int
	Price        int
	Size         int // Size of the largest offer
	Total        int // Total size of all offers
}

// Return the list of open bids (price desc, quantity desc) and open asks (ordered asc, quantity desc)
func GetBidsAsks(allOrders []Order) (bids []Order, asks []Order, bidAggr Aggregates, askAggr Aggregates, err error) {
	// seperate the orders into bids and asks
	for _, o := range allOrders {
		if o.AskBidInd == "BID" {
			bids = append(bids, o)
		} else {
			asks = append(asks, o)
		}
	}

	// sort bids descending price
	sort.Sort(sort.Reverse(sortAscendingPrice(bids)))

	// sort asks ascending price
	sort.Sort(sortAscendingPrice(asks))

	// TODO: Calculate aggregates
	// fmt.Printf("[%s]- Bids: %v\n", time.Now().Format("20060102150405"), bids)
	// fmt.Printf("[%s]- Asks: %v\n", time.Now().Format("20060102150405"), asks)

	return bids, asks, bidAggr, askAggr, err
}

// Get the bid with the highest price
func GetHighestBid(bids []Order) (o Order, err error) {
	if len(bids) <= 0 {
		return o, fmt.Errorf("There are no bids.")
	}
	sort.Sort(sort.Reverse(sortAscendingPrice(bids)))
	return bids[0], nil
}

// Get the ask with the lowest price
func GetLowestAsk(asks []Order) (o Order, err error) {
	if len(asks) <= 0 {
		return o, fmt.Errorf("There are no asks.")
	}
	sort.Sort(sortAscendingPrice(asks))
	return asks[0], nil
}

// get contracts to accept
func GetContractsToAccept(orderbook []Order) (c []int, err error) {
	// get bids and asks
	bids, asks, _, _, err := GetBidsAsks(orderbook)
	handleError(err)

	// find contract to sell
	bid, err := GetHighestBid(bids)
	if err != nil {
		if err.Error() == "There are no bids." {
			return c, fmt.Errorf("Nothing to accept.")
		} else {
			handleError(err)
		}
	}

	// find contract to buy
	ask, err := GetLowestAsk(asks)
	if err != nil {
		if err.Error() == "There are no asks." {
			return c, fmt.Errorf("Nothing to accept.")
		} else {
			handleError(err)
		}
	}

	// return contract to buy and contract to sell
	c = append(c, ask.ContractIdx)
	c = append(c, bid.ContractIdx)
	return c, nil
}

// If there is an error then panic
func handleError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// Function used to sort orders at an sortAscendingPrice
type sortAscendingPrice []Order

func (v sortAscendingPrice) Len() int {
	return len(v)
}

func (v sortAscendingPrice) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v sortAscendingPrice) Less(i, j int) bool {
	return v[i].Price < v[j].Price
}

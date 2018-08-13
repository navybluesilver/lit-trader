package counterparty

import (
	"fmt"
)

type Counterparty struct {
	Name			string //User friendly name
	LNAddress string
	IP 				string
	Port   		uint32
	URL       string //can be blank, but could be used to derive Alias, LNAddress, IP, Port
}

func (c *Counterparty) GetConnectionString() string {
	return fmt.Sprintf("%s@%s:%s", c.LNAddress, c.IP, c.Port)
}

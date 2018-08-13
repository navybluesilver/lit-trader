package trader

import (
	"encoding/json"
	"net/http"
	"fmt"
	config "github.com/navybluesilver/lit-trader/config"
)

var (
		oracleUrl      string = config.GetString("oracle.url")
		oracleName     string = config.GetString("oracle.name")
		datasourceId   int    = config.GetInt("oracle.datasource_id")
)

type Rpoint struct {
	R string `json:"R"`
}

type Pubkey struct {
	A string `json:"A"`
}

type OracleSignature struct {
	Signature string `json:"signature"`
	Value     int    `json:"value"`
}

type OracleDatasources []struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	ID           int    `json:"id"`
	CurrentValue int    `json:"currentValue"`
}

// Check if the oracle exists
func (t *Trader) oracleExists() (bool, error) {
	allOracles, err := t.Lit.ListOracles()
	handleError(err)

	for _, o := range allOracles {
		if o.Name == oracleName { //TODO: check based on pubkey instead
			return true, nil
		}
	}
	return false, nil
}

// Return the oracle index
func (t *Trader) getOracleIdx(oracleName string) (uint64, error) {
	allOracles, err := t.Lit.ListOracles()
	handleError(err)

	for _, o := range allOracles {
		if o.Name == oracleName {
			return o.Idx, nil
		}
	}
	return 0, fmt.Errorf("Oracle [%s] not found", oracleName)
}

func (m *Trader) GetOraclePubKey() (string) {
	var pubkey Pubkey
	url := fmt.Sprintf("%s/api/pubkey", oracleUrl)
	req, err := http.NewRequest("GET", url, nil)
	handleError(err)

	client := &http.Client{}

	resp, err := client.Do(req)
	handleError(err)

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&pubkey)
	handleError(err)

	return pubkey.A
}

func GetR(timestamp int) string {
	var r Rpoint
	url := fmt.Sprintf("%s/api/rpoint/2/%d", oracleUrl, timestamp)
	req, err := http.NewRequest("GET", url , nil)
	handleError(err)

	client := &http.Client{}

	resp, err := client.Do(req)
	handleError(err)

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&r)
	handleError(err)

	return r.R
}

func GetOracleSignature() (oracleValue int64, oracleSignature []byte) {
	var sig OracleSignature
	url := fmt.Sprintf("%s/api/publication/%s", oracleUrl, GetR(GetSettlementTime()))

	req, err := http.NewRequest("GET", url, nil)
	handleError(err)

	client := &http.Client{}

	resp, err := client.Do(req)
	handleError(err)

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&sig)
	handleError(err)

	return int64(sig.Value), []byte(sig.Signature)
}

func (t *Trader) GetCurrentSpot() (int) {
	var datasources OracleDatasources
	url := fmt.Sprintf("%s/api/datasources", oracleUrl)

	req, err := http.NewRequest("GET", url, nil)
	handleError(err)

	client := &http.Client{}

	resp, err := client.Do(req)
	handleError(err)

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&datasources)
	handleError(err)

	for _, datasource := range datasources {
    if datasource.ID == datasourceId {
			return datasource.CurrentValue
		}
	}

	return 0
}

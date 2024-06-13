package x

import (
	"encoding/json"
	"fmt"
	yaml_v3 "gopkg.in/yaml.v3"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func FetchPrices() ([]DenomPrice, error) {
	osmosisRPC := os.Getenv("OSMOSIS_RPC")

	var denomPrices []DenomPrice

	for _, vaultAddress := range Vaults {
		fmt.Println("Fetching prices for vault", vaultAddress)
		// Construct and execute the command
		cmdStr := fmt.Sprintf("osmosisd q concentratedliquidity user-positions %s --node %s", vaultAddress, osmosisRPC)
		fmt.Println("Executing command:", cmdStr)
		cmdArgs := strings.Fields(cmdStr)
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		output, err := cmd.Output()
		fmt.Println("Output:", string(output))
		if err != nil {
			return nil, fmt.Errorf("error executing command for vault %s: %s", vaultAddress, err)
		}

		// Unmarshal the output
		var positions OsmosisPositionsResponse
		err = yaml_v3.Unmarshal(output, &positions)
		if err != nil {
			return nil, fmt.Errorf("error parsing YAML output for vault %s: %s", vaultAddress, err)
		}

		// Process and fetch prices
		for _, position := range positions.Positions {
			denominations := []string{position.Asset0.Denom, position.Asset1.Denom}
			for _, denom := range denominations {
				if !isDenomPresent(denomPrices, denom) {
					price, err := getPriceFromAPI(denom)
					if err != nil {
						return nil, err
					}
					denomPrices = append(denomPrices, DenomPrice{Denom: denom, Price: price})
				}
			}
		}

		// Sleep to avoid rate limits
		fmt.Println("Sleeping for 5 seconds...")
		time.Sleep(5 * time.Second)
	}

	// Save to JSON file
	err := savePricesToFile(denomPrices)
	if err != nil {
		return nil, err
	}

	return denomPrices, nil
}

func isDenomPresent(prices []DenomPrice, denom string) bool {
	for _, price := range prices {
		if price.Denom == denom {
			return true
		}
	}
	return false
}

func getPriceFromAPI(denom string) (float64, error) {
	denomUSDC := "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4"
	if denom == denomUSDC {
		return 1, nil
	}

	amount := 1000000 // 1 * 10^6 (micro unit)
	if decimals, ok := CustomDecimals[denom]; ok {
		// Adjust the amount by the custom decimal difference
		amount *= int(math.Pow10(decimals - 6))
	}

	apiURL := fmt.Sprintf("https://sqs.osmosis.zone/router/quote?tokenIn=%d%s&tokenOutDenom=%s", amount, denom, denomUSDC)
	resp, err := http.Get(apiURL)
	if err != nil {
		return 0, fmt.Errorf("error making request to API: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response body: %s", err)
	}

	var priceResponse DenomPriceResponse
	err = json.Unmarshal(body, &priceResponse)
	if err != nil {
		return 0, fmt.Errorf("error unmarshalling JSON: %s", err)
	}

	price, err := strconv.ParseFloat(priceResponse.AmountOut, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing price: %s", err)
	}

	// Adjusting for micro unit (1e6) or custom decimal places
	price /= math.Pow10(6)
	return price, nil
}

func savePricesToFile(prices []DenomPrice) error {
	fileName := fmt.Sprintf("prices/%s.json", time.Now().Format("01-02-2006"))
	fileData, err := json.Marshal(prices)
	if err != nil {
		return err
	}

	// Ensure the prices directory exists
	if err := os.MkdirAll("prices", 0755); err != nil {
		return err
	}

	return os.WriteFile(fileName, fileData, 0644)
}

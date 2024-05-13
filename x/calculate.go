package x

import (
	sdkmath "cosmossdk.io/math"
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	yaml_v3 "gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var lut = []GrantRange{
	{Min: 0, Max: 250000, Reward: 277778},
	{Min: 250000, Max: 500000, Reward: 308744},
	{Min: 500000, Max: 1000000, Reward: 339784},
	{Min: 1000000, Max: 1750000, Reward: 374312},
	{Min: 1750000, Max: 2750000, Reward: 412807},
	{Min: 2750000, Max: 4000000, Reward: 456058},
	{Min: 4000000, Max: 5500000, Reward: 505263},
	{Min: 5500000, Max: 7250000, Reward: 562178},
	{Min: 7250000, Max: 9250000, Reward: 629365},
	{Min: 9250000, Max: 12000000, Reward: 710558},
	{Min: 12000000, Max: 1000000000, Reward: 833333}, // Max 1B to simulate Infinity
}

func GetVaultsTvl(blockHeight uint64) (VaultsTvlResponse, error) {
	osmosisRPC := os.Getenv("OSMOSIS_RPC")
	blockDate, err := getBlockFromAPI(blockHeight)
	if err != nil {
		return VaultsTvlResponse{}, err
	}

	var vaultTvls []VaultTvl
	var totalTvl float64

	for _, vaultAddress := range Vaults {
		// Construct and execute the command (this could be removed if we find a public archival node with lcd rest api endpoints, and if they support --height flag)
		cmdStr := fmt.Sprintf("osmosisd q concentratedliquidity user-positions %s --node %s --height %d", vaultAddress, osmosisRPC, blockHeight)
		fmt.Println(cmdStr)
		cmdArgs := strings.Fields(cmdStr)
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		output, err := cmd.Output()
		if err != nil {
			return VaultsTvlResponse{}, fmt.Errorf("error executing command for vault %s: %s", vaultAddress, err)
		}

		// Unmarshal the output
		var positions OsmosisPositionsResponse
		err = yaml_v3.Unmarshal(output, &positions)
		if err != nil {
			return VaultsTvlResponse{}, fmt.Errorf("error parsing YAML output for vault %s: %s", vaultAddress, err)
		}

		// Process and populate VaultTvl
		var totalUsd float64
		var tvl VaultTvl
		tvl.Address = vaultAddress
		for _, position := range positions.Positions {
			amount0, ok := new(big.Int).SetString(position.Asset0.Amount, 10)
			if !ok {
				return VaultsTvlResponse{}, fmt.Errorf("error converting amount0 to big.Int for vault %s", vaultAddress)
			}
			amount1, ok := new(big.Int).SetString(position.Asset1.Amount, 10)
			if !ok {
				return VaultsTvlResponse{}, fmt.Errorf("error converting amount1 to big.Int for vault %s", vaultAddress)
			}

			asset0 := sdk.NewCoin(position.Asset0.Denom, sdkmath.NewIntFromBigInt(amount0))
			asset1 := sdk.NewCoin(position.Asset1.Denom, sdkmath.NewIntFromBigInt(amount1))

			tvl.Coins = append(tvl.Coins, asset0, asset1) // append coins to empty array
			totalUsd += convertDenomAmountToUsd(blockDate, asset0) + convertDenomAmountToUsd(blockDate, asset1)
		}
		tvl.TotalUsd = fmt.Sprintf("%.6f", totalUsd) // Format to a fixed number of decimal places
		totalTvl += totalUsd
		vaultTvls = append(vaultTvls, tvl)

		// Sleeps for 1 second avoiding RPCs rate limits
		time.Sleep(2 * time.Second)
	}

	return VaultsTvlResponse{
		Single:   vaultTvls,
		TotalUsd: totalTvl,
	}, nil
}

func getBlockFromAPI(blockHeight uint64) (string, error) {
	apiURL := fmt.Sprintf("https://lcd.osmosis.zone/cosmos/base/tendermint/v1beta1/blocks/%d", blockHeight)
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("error making request to API: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %s", err)
	}

	var blockResponse BlockResponse
	err = json.Unmarshal(body, &blockResponse)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling JSON: %s", err)
	}

	// Parse the time to time.Time object
	parsedTime, err := time.Parse(time.RFC3339, blockResponse.Block.Header.Time)
	if err != nil {
		return "", fmt.Errorf("error parsing block time: %s", err)
	}

	// Format the time to mm-dd-yyyy string
	date := parsedTime.Format("01-02-2006")

	return date, nil
}

func convertDenomAmountToUsd(blockDate string, coin sdk.Coin) float64 {
	// Construct the file path for the prices JSON file
	pricesFilePath := fmt.Sprintf("prices/%s.json", blockDate)

	// Read the JSON file
	fileData, err := os.ReadFile(pricesFilePath)
	if err != nil {
		log.Fatalf("Error reading prices file: %s", err)
		return 0
	}

	// Unmarshal the JSON data into an array of DenomPrice
	var prices []DenomPrice
	err = json.Unmarshal(fileData, &prices)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON data: %s", err)
		return 0
	}

	decimalFactor := 1e6
	if customDecimal, ok := CustomDecimals[coin.Denom]; ok {
		decimalFactor = math.Pow10(customDecimal)
	}

	// Convert math.Int to float64 correctly
	coinAmount := new(big.Float).SetInt(coin.Amount.BigInt())
	coinAmount = coinAmount.Quo(coinAmount, new(big.Float).SetFloat64(decimalFactor))
	coinAmountFloat, _ := coinAmount.Float64()

	// Find the matching denomination and return its price
	for _, price := range prices {
		if price.Denom == coin.Denom {
			return coinAmountFloat * price.Price
		}
	}

	return 0
}

func CalculateGrant(totalTvl float64) float64 {
	for _, lutRange := range lut {
		if totalTvl >= lutRange.Min && totalTvl < lutRange.Max {
			return lutRange.Reward
		}
	}

	return 0
}

package main

import (
	"calculate_grant/x"
	"fmt"
	"github.com/spf13/cobra"
	yaml_v3 "gopkg.in/yaml.v3"
	"log"
	"os"
	"strconv"
)

func main() {
	var rootCmd = &cobra.Command{Use: "app"}

	var cmdFetchPrices = &cobra.Command{
		Use:   "fetch-prices",
		Short: "Retrieves current market prices for all vaults denominations.",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := x.FetchPrices()
			if err != nil {
				fmt.Printf("Error fetching prices: %s\n", err.Error())
				return
			}

			fmt.Println("Prices successfully fetched and saved into prices folder.")
		},
	}

	var cmdCalculateGrant = &cobra.Command{
		Use:   "calculate-grant [block_height]",
		Short: "Calculates the total grant amount based on the Total Value Locked (TVL) at a specific blockchain height",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			blockHeight := args[0]
			blockHeightUint, err := strconv.ParseUint(blockHeight, 10, 64)
			if err != nil {
				fmt.Println("Invalid block height argument. Block height must be an unsigned integer.")
				return
			}

			vaultsTvl, err := x.GetVaultsTvl(blockHeightUint)
			if err != nil {
				fmt.Printf("Error getting vaults tvl: %s\n", err.Error())
				return
			}

			// Build response as ResponseObject
			response := x.ResponseObject{
				TvlPerVault: vaultsTvl.Single,
				TotalTvl:    fmt.Sprintf("%.6f", vaultsTvl.TotalUsd), // Format to a fixed number of decimal places
				Grant:       x.CalculateGrant(vaultsTvl.TotalUsd),
			}

			output, err := yaml_v3.Marshal(response)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(string(output))
		},
	}

	rootCmd.AddCommand(cmdFetchPrices, cmdCalculateGrant)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

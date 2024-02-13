package x

import sdk "github.com/cosmos/cosmos-sdk/types"

type GrantRange struct {
	Min    float64
	Max    float64
	Reward float64
}

type DenomPrice struct {
	Denom string
	Price float64
}

type DenomPriceResponse struct {
	AmountOut string `json:"amount_out"`
}

type BlockResponse struct {
	Block struct {
		Header struct {
			Time string `yaml:"time"`
		} `json:"header"`
	} `json:"block"`
}

type VaultTvl struct {
	Address  string
	Coins    []sdk.Coin
	TotalUsd string
}

type VaultsTvlResponse struct {
	Single   []VaultTvl
	TotalUsd float64
}

type OsmosisPositionsResponse struct {
	Positions []struct {
		Asset0 struct {
			Amount string `yaml:"amount"`
			Denom  string `yaml:"denom"`
		} `yaml:"asset0"`
		Asset1 struct {
			Amount string `yaml:"amount"`
			Denom  string `yaml:"denom"`
		} `yaml:"asset1"`
	} `yaml:"positions"`
}

type ResponseObject struct {
	TvlPerVault []VaultTvl
	TotalTvl    string
	Grant       float64
}

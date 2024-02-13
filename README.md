# Strategist Grant Calculator

## Introduction

The Strategist Grant Calculator is a CLI tool designed to provide transparent and accurate financial calculations for
blockchain protocol governance. This tool fetches data directly from the blockchain and utilizes the Osmosis SQS router
for price conversions, ensuring decisions are based on real, verifiable on-chain values. These external sources are
chosen for their reliability and ease of verification, allowing users to
independently confirm that the prices used in calculations are correct.

## Requirements

Before using the Strategist Grant Calculator, you must have the following installed:

- **Go:** The programming language used to build the tool. [Download and install Go](https://golang.org/dl/).
- **Osmosis Archival Node:** For querying specific block heights in the past. You will need access to an Osmosis
  archival node. This can be a node you run yourself, or a publicly available node.

## Setting up the Environment

### OSMOSIS_RPC Environment Variable

To query a specific block height, the tool requires access to an Osmosis archival node. You need to set
the `OSMOSIS_RPC` environment variable to the address of the node. This can be done in the following ways:

**Using Export:** In your terminal, you can set the variable for your current session with the following command:

```bash
export OSMOSIS_RPC="https://your-archival-node-address:port"
```

**Permanent Setup:** To permanently set the environment variable, you can add the export command to your shell's profile
script (like `~/.bashrc` or `~/.zshrc` for Bash and Zsh respectively).

```bash
echo 'export OSMOSIS_RPC="https://your-archival-node-address:port"' >> ~/.bashrc
```

Then, reload your profile with `source ~/.bashrc` or open a new terminal session.

## Usage

### Fetching Prices

The fetch-prices command is scheduled to run daily via cron or CI to update prices. It fetches the latest price data and
saves it in the prices folder for later use.

```bash
go run main.go fetch-prices
```

### Calculate the Grant Amount

This command will output the total TVL per vault, the total TVL, and the strategist grant amount calculated based on the
specified block height. The tool ensures that all the data is fetched from the blockchain for accuracy and transparency.

```bash
go run main.go calculate-grant <block_height>
```

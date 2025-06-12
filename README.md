# crypto-pb

crypto-pb is an extension for [PocketBase](https://pocketbase.io/) that automatically
scans stored wallet addresses for transactions. It currently supports BTC
(mainnet and testnet) and stores the results in PocketBase collections.

The goal is to integrate crypto transactions into your backend flow without the
need to poll blockchains yourself.

## PocketBase version

The project targets **PocketBase 0.22.x** (tested with 0.22.7). Other versions may
work but are not verified.

## Status / Roadmap

* Store the number of confirmations received
* Proper ETH support
* Assign users to wallets and transactions so only they can read them

## Usage

Build and run the application just like any PocketBase app:

```bash
go run main.go
```

The application will ensure the required collections exist and will start a
scheduler that periodically scans all stored wallets.

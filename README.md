# crypto-pb

crypto-pb is an extension for [PocketBase](https://pocketbase.io/) that automatically
scans stored wallet addresses for transactions. It currently supports BTC
(mainnet and testnet) and stores the results in PocketBase collections.

The goal is to integrate crypto transactions into your backend flow without the
need to poll blockchains yourself.

## Version

Current project version: **0.1**

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

Build and run a binary:

```bash
go build -o crypto-pb
# run against mainnet
./crypto-pb serve
# run against testnet
./crypto-pb serve --testnet
```

If `--testnet` is omitted the application uses Bitcoin mainnet. The application will ensure the required collections exist and will start a
scheduler that periodically scans all stored wallets.

## Python example

The `examples/python/client.py` script demonstrates how to authenticate with the REST API and retrieve data:

1. Login as a user
2. List wallets with their address and balance
3. List all transactions for a selected wallet (amount and confirmations)

List all wallets for a user:
```bash
python examples/python/client.py --email user@example.com --password secret
```

List all transactions for a specific wallet owned by that user:
```bash
python examples/python/client.py --email user@example.com --password secret --wallet <address>
```

## Contributing

* **Feature requests**: please open an issue in the repository.
* **Code contributions** are very welcome. Fork the repo and open a pull request.
* If you find the project useful consider supporting it via GitHub's **Support** button.

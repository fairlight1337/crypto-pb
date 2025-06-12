import argparse
import requests


def main():
    parser = argparse.ArgumentParser(description="crypto-pb demo client")
    parser.add_argument("--host", default="http://127.0.0.1:8090", help="PocketBase host")
    parser.add_argument("--email", required=True, help="user email")
    parser.add_argument("--password", required=True, help="user password")
    parser.add_argument("--wallet", help="wallet id to fetch transactions for")
    args = parser.parse_args()

    # login
    r = requests.post(f"{args.host}/api/collections/users/auth-with-password", json={"identity": args.email, "password": args.password})
    r.raise_for_status()
    token = r.json()["token"]
    headers = {"Authorization": token}

    # get wallets
    resp = requests.get(f"{args.host}/api/collections/cryptowallets/records", headers=headers)
    resp.raise_for_status()
    wallets = resp.json().get("items", [])
    print("Wallets:")
    for w in wallets:
        print(f"{w['address']}, balance: {w.get('balance', 0)}")

    if args.wallet:
        tx_resp = requests.get(
            f"{args.host}/api/collections/cryptotransactions/records",
            params={"filter": f"wallet.address='{args.wallet}'", "sort": "-timestamp"},
            headers=headers,
        )
        tx_resp.raise_for_status()
        print("Transactions:")
        for t in tx_resp.json().get("items", []):
            print(
                f"{t.get('txid')}, amount: {t.get('amount')}, conf: {t.get('confirmations')}"
            )


if __name__ == "__main__":
    main()

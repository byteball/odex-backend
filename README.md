# ODEX Backend

[![Chat on discord](https://img.shields.io/discord/534371689996222485.svg?logo=discord)](https://discord.gg/qHHST4e)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT)
![Contributions welcome](https://img.shields.io/badge/contributions-welcome-orange.svg)

# The ODEX Decentralized Exchange

The ODEX decentralized exchange is a hybrid decentralized exchange that aims at bringing together the ease of use of centralized exchanges along with the security and privacy features of decentralized exchanges. Orders are matched through an off-chain orderbook. After orders are matched and signed, the decentralized exchange operator has the sole ability to perform a transaction to the AA. This provides for the best UX as the exchange operator is the only party having to interact directly with the blockchain. Exchange users simply sign orders which are broadcasted then to the orderbook. This design enables users to queue and cancel their orders seamlessly.

Several matchers can operate exchanges based on ODEX technology at the same time. They share their orderbooks and exchange all new orders among themselves, thus improving liquidity for all ODEX exchanges. An order can be submitted through any ODEX exchange, however to be matched, both maker and taker orders have to indicate the same matcher. The exchange that was used to submit the order serves as an affliate and can charge a fee from its users.  Anyone can become a matcher or affiliate, or just watch the orders that are being exchanged among the matchers and detect any possible misbehavior by matchers.

# Getting Started

## Requirements

- **mongoDB** version 3.6 or newer ([installation instructions for ubuntu](https://docs.mongodb.com/manual/tutorial/install-mongodb-on-ubuntu/))
- **rabbitmq** version 3.7.7 or newer ([installation instructions for ubuntu](https://computingforgeeks.com/how-to-install-latest-rabbitmq-server-on-ubuntu-18-04-lts/))
- **golang** latest ([installation instructions for ubuntu](https://github.com/golang/go/wiki/Ubuntu))

## Install

```
go get github.com/byteball/odex-backend
```

## Run
You don't run the backend directly. Run [ODEX wallet](https://github.com/byteball/odex-wallet) and it will launch the backend automatically.


# API Endpoints

## Tokens
- `GET /tokens` : returns list of all the tokens from the database
- `GET /tokens/<asset>`: returns details of a token from db using token's asset ID
- `POST /tokens`: Create/Insert token in DB. Sample input:
```
{
	"name":"HotPotCoin",
	"symbol":"HPC",
	"decimal":18,
	"asset":"0x1888a8db0b7db59413ce07150b3373972bf818d3",
	"active":true,
	"quote":true
}
```

## Pairs
- `GET /pairs` : returns list of all the pairs from the database
- `GET /pairs/<baseToken>/<quoteToken>`: returns details of a pair from db using using asset IDs of its constituting tokens
- `GET /pairs/book/<pairName>`: Returns orderbook for the pair using pair name
- `POST /pairs/create`: Create/Insert pair in DB. Sample input:
```
{
    "asset":"5b3e82587b44576ba8000001",
}

```

## Address
- `POST /account/create?address=<addr>`: Create/Insert address and corresponding balance entry in DB.

## Balance
- `GET /account/balances/<addr>`: Fetch the balance details from db of the given address.

## Order
- `GET /orders?address=<addr>`: Fetch all the orders placed by the given address

## Trade
- `GET /trades/history/<pair>`: Fetch complete trade history of given pair using pair name
- `GET /trades?address=<addr>`: Fetch all the trades in which the given address is either maker or taker
- `GET /trades/ticks`: Fetch ohlcv data. Query Params:
```
// Query Params for /trades/ticks
pairName: names of pair separated by comma(,) ex: "hpc/aut,abc/xyz". (At least 1 Required)
unit: sec,min,hour,day,week,month,yr. (default:hour)
duration: in int. (default: 24)
from: unix timestamp of from time.(default: start of timestamp)
to: unix timestamp of to time. (default: current timestamp)
```

# Types

## Orders

Orders contain the information that is required to register an order in the orderbook as a "Maker".

- **id** is the primary ID of the order (possibly deprecated)
- **orderType** is either BUY or SELL. It is currently not parsed by the server and compute directly from tokenBuy, tokenSell, amountBuy, amountSell
- **exchangeAddress** is the exchange AA address
- **maker** is the maker (usually sender) Obyte address
- **tokenBuy** is the BUY token asset ID
- **tokenSell** is the SELL token asset ID
- **amountBuy** is the BUY amount (in BUY_TOKEN units)
- **amountSell** is the SELL amount (in SELL_TOKEN units)
- **expires** is the order expiration timestamp
- **nonce** is a random string or number to make sure order hashes are unique even if all other parameters are identical
- **pairID** is a hash of the corresponding
- **hash** is a hash of the order details (see details below)
- **signature** is a signature of the order hash. The signer must equal to the maker address for the order to be valid.
- **price** corresponds to the pricepoint computed by the matching engine (not parsed)
- **amount** corresponds to the amount computed by the matching engine (not parsed)

**Order Price and Amount**

There are two ways to describe the amount of tokens being bought/sold. The AA requires (sell_asset, buy_asset, sell_amount, price) while the
orderbook requires (pairID, amount, price).

The conversion between both systems can be found in the engine.ComputeOrderPrice
function


**Order Hash**

The order hash is a sha-256 hash of the following elements:
- Exchange address
- Token Buy asset ID
- Amount Buy
- Token Sell asset ID
- Amount Sell
- Nonce
- User Address


## Trades


- **orderHash** is the hash of the matching order
- **amount** is the amount of tokens that will be traded
- **taker** is the taker Obyte address
- **pairID** is a hash identifying the token pair that will be traded
- **hash** is a unique identifier hash of the trade details (see details below)

Trade Hash:

The trade hash is a sha-256 hash of the following elements:
- Order Hash
- Amount
- Taker Address


The (Order, Trade) tuple can then be used to perform an on-chain transaction for this trade.

## Quote Tokens and Token Pairs

In the same way as traditional exchanges function with the idea of base
currencies and quote currencies, the ODEX decentralized exchange works with
base tokens and quote tokens under the following principles:

- Only the exchange operator can register a quote token
- Anybody can register a token pair (but the quote token needs to be registered)

Token pairs are identified by an ID (a hash of both token asset IDs)



# Websocket API

See [WEBSOCKET_API.md](WEBSOCKET_API.md)


# Contribution

Thank you for considering helping the ODEX project !

To make the ODEX project truely revolutionary, we need and accept contributions from anyone and are grateful even for the smallest fixes.

If you want to help ODEX, please fork and setup the development environment of the appropriate repository. In the case you want to submit substantial changes, please get in touch with our development team on #odex channel on [Obyte Discord](https://discord.obyte.org/) to verify those modifications are in line with the general goal of the project and receive early feedback. Otherwise you are welcome to fix, commit and send a pull request for the maintainers to review and merge into the main code base.

Please make sure your contributions adhere to our coding guidelines:

Code must adhere as much as possible to standard conventions (DRY - Separation of concerns - Modular)
Pull requests need to be based and opened against the master branch
Commit messages should properly describe the code modified
Ensure all tests are passing before submitting a pull request

# Contact

If you have questions, ideas or suggestions, you can reach the development team on Discord in the #odex channel.  [Discord Link](https://discord.obyte.org/)

## Credits

ODEX backend is based on [AMP Exchange](https://github.com/Proofsuite/amp-matching-engine), the most beautiful and easy to use decentralized exchange.

# License

All the code in this repository is licensed under the MIT License, also included in our repository in the LICENSE file.

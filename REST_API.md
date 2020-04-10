# REST API

There are 7 different resources on the matching engine REST API:

* accounts
* pairs
* tokens
* trades
* orderbook
* orders
* ohlcv


# Account resource

### GET /account/{userAddress}

Retrieve the account information for a certain Obyte address (mainly token balances)

### GET /account/{userAddress}/{asset}

Retrieve the token balance of a certain Obyte address

* {userAddress} is the Obyte address of a user/client wallet
* {asset} is the ID of an asset (base or quote)


# Pairs resource

### GET /pair?baseToken={baseToken}&quoteToken={quoteToken}

Retrieve the pair information corresponding to a baseToken and a quoteToken where:

* {baseToken} is the ID of a base token
* {quoteToken} is the ID of a quote token

### GET /pairs

Retrieve all pairs currently registered on the exchange

### GET /pairs/data?baseToken={baseToken}&quoteToken={quoteToken}

Retrieve pair data corresponding to a baseToken and quoteToken where

* {baseToken} is the ID of a base token
* {quoteToken} is the ID of a quote token

This endpoints returns the Open, High, Low, Close, Volume and Change for the last 24 hours
as well as the last price.


# Tokens resource

### GET /tokens

Retrieve all tokens currently registered on the exchange

### GET /tokens/base

Retrieve all base tokens currently registered on the exchange

### GET /tokens/quote

Retrieve all quote tokens currently registered on the exchange

### GET /tokens/{assetOrSymbol}

Retrieve token information for a token asset ID or symbol

* {assetOrSymbol} is an asset ID ("base" or 44-bytes long string) or symbol


# Orderbook resource

### GET /orderbook?baseToken={baseToken}&quoteToken={quoteToken}

Retrieve the orderbook (amount and pricepoint) corresponding to a a baseToken and a quoteToken where:

* {baseToken} is the ID of a base token
* {quoteToken} is the ID of a quote token

### GET /orderbook/raw?baseToken={baseToken}&quoteToken={quoteToken}

Retrieve the orderbook (full raw orders, including fields such as hashes, maker, taker addresses, signatures, etc.)
corresponding to a baseToken and a quoteToken.

* {baseToken} is the ID of a base token
* {quoteToken} is the ID of a quote token


# Trade resource

### GET /trades?address={address}

Retrieve the sorted list of trades for an Obyte address

* {address} is an Obyte address

### GET /trades/pair?baseToken={baseToken}&quoteToken={quoteToken}

Retrieve all trades corresponding to a baseToken and a quoteToken

* {baseToken} is the ID of a base token
* {quoteToken} is the ID of a quote token



# Order resource

### GET /orders?address={address}

Retrieve the sorted list of orders for an Obyte address

### GET /orders/positions?address={address}

Retrieve the list of positions for an Obyte address. Positions are order that have been sent
to the matching engine and that are waiting to be matched

* {address} is an Obyte address

### GET /orders/history?address={address}

Retrieve the list of filled order for an Obyte address.

* {address} is an Obyte address


# OHLCV resource

### GET /ohlcv?baseToken={baseToken}&quoteToken={quoteToken}&pairName={pairName}&unit={unit}&duration={duration}&from={from}&to={to}

Retrieve OHLCV data corresponding to a baseToken and a quoteToken.

* {baseToken} is the ID of a baseToken
* {quoteToken} is the ID of a quoteToken
* {pairName} is the pair name under the format {baseTokenSymbol}/{quoteTokenSymbol}(eg. "ZRX/WETH"). I believe this parameter is currently required but it's planned to be optional. The idea is for this parameter to be used for verifications purposes and the API to send back an error if it does not correspond to a baseToken/quoteToken parameters
* {duration} is the duration (in units, see param below) of each candlestick
* {units} is the unit used to represent the above duration: "minute", "hour", "day", "week", "month"
* {from} is the beginning timestamp from which ohlcv data has to be queried
* {to} is the ending timestamp until which ohlcv data has to be queried

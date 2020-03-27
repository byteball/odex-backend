# Websocket API

**Websocket Endpoint**: `/socket`

There are 5 channels on the matching engine websocket API:

* orders
* ohlcv
* orderbook
* raw_orderbook
* trades

To send a message to a specific channel, the general format of a message is the following:

```json
{
  "channel": <channel_name>,
  "event": {
    "type": <event_type>,
    "payload": <payload>
  }
}
```

where

* \<channel_name> is either 'orders', 'ohlcv', 'orderbook', 'raw_orderbook', 'trades'
* \<event_type> is a string describing what type of message is being sent
* \<payload> is a JSON object


# Trades Channel

## Message:

* SUBSCRIBE_TRADES (client --> server)
* UNSUBSCRIBE_TRADES (client --> server)
* INIT (server --> client)
* UPDATE (server --> client)

## SUBSCRIBE_TRADES MESSAGE (client --> server)

```json
{
  "channel": "trades",
  "event": {
    "type": "SUBSCRIBE",
    "payload": {
      "baseToken": <asset>,
      "quoteToken": <asset>,
      "name": <baseTokenSymbol>/<quoteTokenSymbol>,
    }
  }
}
```

The name parameter is optional but the API will return an error if the symbols do not correspond to the symbols registered
in the matching engine. This optional parameter can be used for verifying you are subscribing to the right channel.

###Example:
```json
{
  "channel": "trades",
  "event": {
    "type": "SUBSCRIBE",
    "payload": {
      "baseToken": "0x546d3B3d69E30859f4F3bA15F81809a2efCE6e67",
      "quoteToken": "0x17b4E8B709ca82ABF89E172366b151c72DF9C62E",
      "name": "FUN/WETH",
    }
  }
}
```


## UNSUBSCRIBE_TRADES MESSAGE (client --> server)

```json
{
  "channel": "trades",
  "event": {
    "type": "UNSUBSCRIBE",
  }
}
```

## INIT MESSAGE (server --> client)

The general format of the init message is the following:

```json
{
  "channel": "trades",
  "event": {
    "type": "INIT",
    "payload": [
      <trade>,
      <trade>,
      ...
    ]
  }
}
```

## UPDATE MESSAGE (server --> client)

The general format of the update message is the following:

```json
{
  "channel": "trades",
  "event": {
    "type": "UPDATE",
    "payload": [
      <trade>,
      <trade>,
      ...
    ]
  }
}
```

The format of the message is identical to the INIT message.
This message differs from the INIT message in that the client is supposed
to update trades with identical hashes in their data set. The INIT messages
suppose that there are no other existing trades for the currently subscribed pair.





# Orderbook Channel

## Message:
* SUBSCRIBE_ORDERBOOK (client --> server)
* UNSUBSCRIBE_ORDERBOOK (client --> server)
* INIT (server --> client)
* UPDATE (server --> client)


## SUBSCRIBE_ORDERBOOK MESSAGE (client --> server)

```json
{
  "channel": "orderbook",
  "event": {
    "type": "SUBSCRIBE",
    "payload": {
      "baseToken": <asset>,
      "quoteToken": <asset>,
      "name": <baseTokenSymbol>/<quoteTokenSymbol>,
    }
  }
}
```

### Example:

```json
{
  "channel": "orderbook",
  "event": {
    "type": "SUBSCRIBE",
    "payload": {
      "baseToken": "0x546d3B3d69E30859f4F3bA15F81809a2efCE6e67",
      "quoteToken": "0x17b4E8B709ca82ABF89E172366b151c72DF9C62E",
      "name": "FUN/WETH",
    }
  }
}
```

## UNSUBSCRIBE_ORDERBOOK MESSAGE (client --> server)

```json
{
  "channel": "orderbook",
  "event": {
    "type": "UNSUBSCRIBE",
  }
}
```


## INIT MESSAGE (server --> client)

The general format of the INIT message is the following:

```json
{
  "channel": "orderbook",
  "event": {
    "type": "INIT",
    "payload": {
      "asks": [ <ask>, <ask>, ... ],
      "bids": [ <bid>, <bid>, ... ],
    }
  }
}
```

# Example:

```json
{
  "channel": "orderbook",
  "event": {
    "type": "UDPATE",
    "payload": {
      "asks": [
        { "amount": "10000", "price": 1.234, "matcherAddress": "SPV5WIBQQT4DMW7UU5GWCMLYDVNGKECD" },
        { "amount": "10000", "price": 1.235, "matcherAddress": "SPV5WIBQQT4DMW7UU5GWCMLYDVNGKECD" },
      ],
      "bids": [
        { "amount": "10000", "price": 1.233, "matcherAddress": "SPV5WIBQQT4DMW7UU5GWCMLYDVNGKECD" },
        { "amount": "10000", "price": 1.232, "matcherAddress": "SPV5WIBQQT4DMW7UU5GWCMLYDVNGKECD" },
      ]
    }
  }
}
```

## UPDATE MESSAGE (server --> client)

The general format of the update message is the following:

```json
{
  "channel": "orderbook",
  "event": {
    "type": "UPDATE",
    "payload": {
      "asks": [ <ask>, <ask>, ... ],
      "bids": [ <bid>, <bid>, ... ],
    },
  },
}
```

The format of the message is identical to the INIT message.


# Example:

```json
{
  "channel": "orderbook",
  "event": {
    "type": "UDPATE",
    "payload": {
      "asks": [
        { "amount": "10000", "price": 1.234, "matcherAddress": "SPV5WIBQQT4DMW7UU5GWCMLYDVNGKECD" },
        { "amount": "10000", "price": 1.235, "matcherAddress": "SPV5WIBQQT4DMW7UU5GWCMLYDVNGKECD" },
      ],
      "bids": [
        { "amount": "10000", "price": 1.233, "matcherAddress": "SPV5WIBQQT4DMW7UU5GWCMLYDVNGKECD" },
        { "amount": "10000", "price": 1.232, "matcherAddress": "SPV5WIBQQT4DMW7UU5GWCMLYDVNGKECD" },
      ]
    }
  }
}
```

# OHLCV Channel

## Message:
* SUBSCRIBE_OHLCV (client --> server)
* UNSUBSCRIBE_OHLCV (client --> server)
* INIT (server --> client)
* UPDATE (server --> client)


## SUBSCRIBE_OHLCV MESSAGE (client --> server)

```json
{
  "channel": "ohlcv",
  "event": {
    "type": "SUBSCRIBE",
    "payload": {
      "baseToken": <baseAsset>,
      "quoteToken": <quoteAsset>,
      "name": <baseTokenSymbol/quoteTokenSymbol>,
      "from": <from>,
      "to": <to>,
      "duration": <duration>,
      "units": <hour>
    }
  }
}
```

where
* \<baseAsset> is the ID of the base token asset,
* \<quoteAsset> is the ID of the quote token asset,
* \<duration> is the duration (in units, see param below) of each candlestick
* \<units> is the unit used to represent the above duration: "minute", "hour", "day", "week", "month"
* \<from> is the beginning timestamp from which ohlcv data has to be queried
* \<to> is the ending timestamp until which ohlcv data has to be queried

### Example:

```json
{
  "channel": "ohlcv",
  "event": {
    "type": "SUBSCRIBE",
    "payload": {
      "baseToken": "0x546d3B3d69E30859f4F3bA15F81809a2efCE6e67",
      "quoteToken": "0x17b4E8B709ca82ABF89E172366b151c72DF9C62E",
      "name": "FUN/WETH",
      "from": 1534746133,
      "to": 1540016533,
      "duration": 1,
      "units": "hour"
    }
  }
}
```




## UNSUBSCRIBE_OHLCV MESSAGE (client --> server)

```json
{
  "channel": "ohlcv",
  "event": {
    "type": "UNSUBSCRIBE",
  }
}
```




# Orders Channel

## Message:
* NEW_ORDER (client --> server)
* ORDER_ADDED (server --> client)
* CANCEL_ORDER (client --> server)
* ORDER_CANCELLED (server --> client) #CANCELLED with two L
* ORDER_MATCHED (server --> client)
* ORDER_PENDING (server --> client)
* ORDER_SUCCESS (server --> client)
* ORDER_ERROR (server --> client)
* ERROR (server --> client)
* UPDATE (server --> client)

## NEW_ORDER (client --> server)

The general format of the NEW_ORDER message is the following:
```json
{
  "channel": "orders",
  "event": {
    "type": "NEW_ORDER",
    "payload": <signed order>,
  }
}
```

where:
* \<signed order> is an order signed by the client

## Example:
```json
{
  "channel": "orders",
  "event": {
    "type": "NEW_ORDER",
    "payload": {
      "version": "2.0",
      "last_ball_unit": "ONWnTWDzacnUS6AwvbofGpniMKIgOO+mCMBngXa4CcY=",
      "authors": [
        {
          "address": "EDMS22PYWN5NE7F34R5CLNTJSVNLLGLS",
          "authentifiers": {
            "r": "MCGW9yLmdmdks2fj6rF/lNKPAUFxb/4V1HL7kmov47xXzfqGw9Y8KrmaRzc0cfpL3gR0myZzaNT+6DZvR1nJgg=="
          }
        }
      ],
      "signed_message": {
        "aa": "EUZ2QB4UHCMGN2D4LEOL5I3DUGP4FDJT",
        "address": "EDMS22PYWN5NE7F34R5CLNTJSVNLLGLS",
        "buy_asset": "base",
        "matcher": "SPV5WIBQQT4DMW7UU5GWCMLYDVNGKECD",
        "matcher_fee": 557,
        "matcher_fee_asset": "UccpQo12uLmufihkzdK7Kcrb5BlHp8GcMrSEA7NVdNw=",
        "price": 48.47309743092583,
        "sell_amount": 557010,
        "sell_asset": "UccpQo12uLmufihkzdK7Kcrb5BlHp8GcMrSEA7NVdNw="
      }
    }
  }
}
```


## ORDER_ADDED MESSAGE (server --> client)

The general format of the ORDER_ADDED message is the following:

```
{
  "channel": "orders",
  "event": {
    "type": "ORDER_ADDED",
    "payload": <order>
  }
}
```

where:

* \<order> is the information about the order as interpreted by the ODEX node, it encloses the original order sent by the client.


```json
{
  "channel": "orders",
  "event": {
    "type": "ORDER_ADDED",
    "payload": {
      "hash": "Z1g1Pevot/v4lukyW8qixZsw9PZFpqD5NPIVHYLopok=",
      "userAddress": "EDMS22PYWN5NE7F34R5CLNTJSVNLLGLS",
      "matcherAddress": "SPV5WIBQQT4DMW7UU5GWCMLYDVNGKECD",
      "affiliateAddress": "",
      "amount": "100",
      "filledAmount": "0",
      "price": 48.47309743092583,
      "side": "SELL",
      "status": "OPEN",
      "pairName": "FUN/WETH",
      "quoteToken": "UccpQo12uLmufihkzdK7Kcrb5BlHp8GcMrSEA7NVdNw=",
      "baseToken": "base",
      "originalOrder": {
        "version": "2.0",
        "last_ball_unit": "ONWnTWDzacnUS6AwvbofGpniMKIgOO+mCMBngXa4CcY=",
        "authors": [
          {
            "address": "EDMS22PYWN5NE7F34R5CLNTJSVNLLGLS",
            "authentifiers": {
              "r": "MCGW9yLmdmdks2fj6rF/lNKPAUFxb/4V1HL7kmov47xXzfqGw9Y8KrmaRzc0cfpL3gR0myZzaNT+6DZvR1nJgg=="
            }
          }
        ],
        "signed_message": {
          "aa": "EUZ2QB4UHCMGN2D4LEOL5I3DUGP4FDJT",
          "address": "EDMS22PYWN5NE7F34R5CLNTJSVNLLGLS",
          "buy_asset": "base",
          "matcher": "SPV5WIBQQT4DMW7UU5GWCMLYDVNGKECD",
          "matcher_fee": 557,
          "matcher_fee_asset": "UccpQo12uLmufihkzdK7Kcrb5BlHp8GcMrSEA7NVdNw=",
          "price": 48.47309743092583,
          "sell_amount": 557010,
          "sell_asset": "UccpQo12uLmufihkzdK7Kcrb5BlHp8GcMrSEA7NVdNw="
        }
      },
      "createdAt": "2018-10-20T15:21:55.119253+09:00"
    }
  }
}
```


## CANCEL_ORDER MESSAGE (client --> server)

The general format of the CANCEL_ORDER message is the following:

```json
{
  "channel": "orders",
  "event": {
    "type": "CANCEL_ORDER",
    "payload": <signed cancel>
  }
}
```

where:

* \<signed cancel> is a signed cancel command. The message being signed is "Cancel order \<order hash>"

Example:

```json
{
  "channel": "orders",
  "event": {
    "type": "CANCEL_ORDER",
    "payload": {
      "version": "2.0",
      "last_ball_unit": "ONWnTWDzacnUS6AwvbofGpniMKIgOO+mCMBngXa4CcY=",
      "authors": [
        {
          "address": "EDMS22PYWN5NE7F34R5CLNTJSVNLLGLS",
          "authentifiers": {
            "r": "MCGW9yLmdmdks2fj6rF/lNKPAUFxb/4V1HL7kmov47xXzfqGw9Y8KrmaRzc0cfpL3gR0myZzaNT+6DZvR1nJgg=="
          }
        }
      ],
      "signed_message": "Cancel order Z1g1Pevot/v4lukyW8qixZsw9PZFpqD5NPIVHYLopok="
    }
  }
}
```



## ORDER_CANCELLED_MESSAGE (client --> server)

The general format of the order cancelled message is the following:

```json
{
  "channel": "orders",
  "event": {
    "type": "ORDER_CANCELLED",
    "payload": <order>,
  }
}
```

The payload is the same as in ORDER_ADDED.

## Example:
```json
{
  "channel": "orders",
  "event": {
    "type": "ORDER_CANCELLED",
    "payload": {
      "hash": "Z1g1Pevot/v4lukyW8qixZsw9PZFpqD5NPIVHYLopok=",
      "userAddress": "EDMS22PYWN5NE7F34R5CLNTJSVNLLGLS",
      "matcherAddress": "SPV5WIBQQT4DMW7UU5GWCMLYDVNGKECD",
      "affiliateAddress": "",
      "amount": "100",
      "filledAmount": "0",
      "price": 48.47309743092583,
      "side": "SELL",
      "status": "OPEN",
      "pairName": "FUN/WETH",
      "quoteToken": "UccpQo12uLmufihkzdK7Kcrb5BlHp8GcMrSEA7NVdNw=",
      "baseToken": "base",
      "originalOrder": {
        "version": "2.0",
        "last_ball_unit": "ONWnTWDzacnUS6AwvbofGpniMKIgOO+mCMBngXa4CcY=",
        "authors": [
          {
            "address": "EDMS22PYWN5NE7F34R5CLNTJSVNLLGLS",
            "authentifiers": {
              "r": "MCGW9yLmdmdks2fj6rF/lNKPAUFxb/4V1HL7kmov47xXzfqGw9Y8KrmaRzc0cfpL3gR0myZzaNT+6DZvR1nJgg=="
            }
          }
        ],
        "signed_message": {
          "aa": "EUZ2QB4UHCMGN2D4LEOL5I3DUGP4FDJT",
          "address": "EDMS22PYWN5NE7F34R5CLNTJSVNLLGLS",
          "buy_asset": "base",
          "matcher": "SPV5WIBQQT4DMW7UU5GWCMLYDVNGKECD",
          "matcher_fee": 557,
          "matcher_fee_asset": "UccpQo12uLmufihkzdK7Kcrb5BlHp8GcMrSEA7NVdNw=",
          "price": 48.47309743092583,
          "sell_amount": 557010,
          "sell_asset": "UccpQo12uLmufihkzdK7Kcrb5BlHp8GcMrSEA7NVdNw="
        }
      },
      "createdAt": "2018-10-20T15:21:55.119253+09:00"
    }
  }
}
```




## ORDER MATCHED MESSAGE (server --> client)

The general format of notification about matched orders is the following:

```
{
  "channel": "orders",
  "event": {
    "type": "ORDER_MATCHED",
    "payload": {
      "takerOrder": <order>,
      "makerOrders": [ <order>, <order>, ...],
      "trades": [ <trade>, <trade>, ...],
    }
  }
}

```

where

* \<order> is an order that was initially sent by a user/client
* \<trade> is a trade created as a result of matching

This message is sent when an order has been matched in the order book and sent to the transaction queue but not sent to Obyte chain for execution yet.



## ORDER PENDING MESSAGE (server --> client)

The order pending message indicates that the order was successfully sent to Obyte chain for execution. This typically happens immediately after ORDER_MATCHED. When ORDER_PENDING is received, the transaction has not triggered AA execution yet but this will invariably happen after some time (usually, a few minutes).

The general format of the order success message is the following:

```json
{
  "channel": "orders",
  "event": {
    "type": "ORDER_PENDING",
    "payload": {
      "takerOrder": <order>,
      "makerOrders": [ <order>, <order>, ...],
      "trades": [ <trade>, <trade>, ...],
    }
  }
}

```

It is identical to the order matched message except that order statuses are different.


## ORDER SUCCESS MESSAGE (server --> client)

The order success message indicates that a previously delayed or replicated order was successfully sent to Obyte chain for execution. It should be treated the same as ORDER_PENDING.

The general format of the order success message is the following:

```json
{
  "channel": "orders",
  "event": {
    "type": "ORDER_SUCCESS",
    "payload": {
      "takerOrder": <order>,
      "makerOrders": [ <order>, <order>, ...],
      "trades": [ <trade>, <trade>, ...],
    }
  }
}

```

It is identical to the order pending message except that order statuses are different.



## ORDER ERROR MESSAGE (server --> client)

The ORDER_ERROR message indicates that an attempt to send a trade transaction to the DAG was rejected or it was sent to the DAG but was later rejected by the exchange AA.

```json
{
  "channel": "orders",
  "event": {
    "type": "ORDER_ERROR",
    "payload": {
      "takerOrder": <order>,
      "makerOrders": [ <order>, <order>, ...],
      "trades": [ <trade>, <trade>, ...],
    }
  }
}

```


It is identical to the order successs message exect that order statuses are different.
The client should usually not receive this message and it can be interpreted as an 'internal server error' (bug in the system rather than a malformed payload or client error)



## ADDRESS MESSAGE (client --> server)

The general format of the ADDRESS message is the following:

```json
{
  "channel": "orders",
  "event": {
    "type": "ADDRESS",
    "payload": <user address>
  }
}
```

where:

* \<user address> is the user address to be watched in the `orders` channel.  This message opens a subscription for ORDER_* events that affect this address.

Example:
```json
{
  "channel": "orders",
  "event": {
    "type": "ADDRESS",
    "payload": "EDMS22PYWN5NE7F34R5CLNTJSVNLLGLS"
  }
}
```


# Raw Orderbook Channel

## Message:
* SUBSCRIBE (client --> server)
* UNSUBSCRIBE (client --> server)
* INIT (server --> client)
* UPDATE (server --> client)


## SUBSCRIBE_RAW_ORDERBOOK MESSAGE (client --> server)

```json
{
  "channel": "raw_orderbook",
  "event": {
    "type": "SUBSCRIBE",
    "payload": {
      "baseToken": <asset>,
      "quoteToken": <asset>,
      "name": <baseTokenSymbol>/<quoteTokenSymbol>,
    }
  }
}
```

### Example:

```json
{
  "channel": "raw_orderbook",
  "event": {
    "type": "SUBSCRIBE",
    "payload": {
      "baseToken": "0x546d3B3d69E30859f4F3bA15F81809a2efCE6e67",
      "quoteToken": "0x17b4E8B709ca82ABF89E172366b151c72DF9C62E",
      "name": "FUN/WETH",
    }
  }
}
```

## UNSUBSRIBE MESSAGE (client --> server)

```json
{
  "channel": "raw_orderbook",
  "event": {
    "type": "UNSUBSCRIBE",
  }
}
```


## INIT MESSAGE (server --> client)

The general format of the INIT message is the following:

```json
{
  "channel": "raw_orderbook",
  "event": {
    "type": "INIT",
    "payload": [
      <order>,
      <order>,
      <order>
    ]
  }
}
```

## UPDATE MESSAGE (server --> client)

The general format of the UPDATE message is the following:

```json
{
  "channel": "raw_orderbook",
  "event": {
    "type": "UPDATE",
    "payload": [
      <order>,
      <order>,
      <order>
    ]
  }
}
```
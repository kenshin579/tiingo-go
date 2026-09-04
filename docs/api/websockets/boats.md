# Real-time & Historical BOATS Overnight WebSocket API Documentation

> 출처: https://www.tiingo.com/documentation/websockets/boats · 생성: 2026-09-04 (tools/gendocs)

## Endpoints

**WebSocket Endpoints**

The BOATS Overnight Real-time Websocket Endpoints are currently in beta. For production use cases, we recommend the [IEX websocket endpoints](https://www.tiingo.com/documentation/websockets/iex) - which these endpoints expand upon.

Access requires the **BOATS Real-time** entitlement — activate it on the [Entitlements page](https://www.tiingo.com/account/billing/add-ons).

```
# Websocket BOATS API Top-of-Book & Last Trade Endpoint
wss://api.tiingo.com/boats
```

## 3.5.1 Overview

Tiingo provides updates via websocket every time the Top-of-book (best bid/offer prices and sizes) change and when a trade is executed.

We obtain BOATS (Blue Ocean ATS) top-of-book and last-sale feeds, normalize them into Tiingo's realtime array format, and forward every validated update to entitled websocket clients with minimal processing. Updates are delivered as a JSON array inside our JSON object.

To further minimize latency, we use bare metal machines for our realtime infrastructure close to the market-data path.

- Trades are when a security/stock was traded on an exchange and includes the price (lastPrice), the volume done (lastSize), and the four raw MEMOIR sale-condition characters from BOATS (Sale Condition 1–4).
- A quote update is when the bid/ask changes, but no trades were done. Tiingo sends a Quote update message via the Websocket if the Top-of-Book values change. In other words, if the best bid price, best bid size, best ask price, or best ask size change, then an update is sent.

The BOATS firehose is an add-on entitlement (BOATS Real-time). Please note the firehose exposes a very high amount of data, in some cases to the nanosecond resolution. Please build your systems cautiously and to scale.

Access to the BOATS firehose requires the **BOATS Real-time** entitlement on your API account. Without that entitlement, subscribe requests to `wss://api.tiingo.com/boats` are rejected.

The public BOATS websocket currently exposes a single threshold level: **thresholdLevel 3**, which streams every validated venue-native quote and trade update Tiingo receives from BOATS (Blue Ocean ATS).

For REST snapshots and overnight OHLC history, see [2.7 REST - BOATS Overnight Real-time](https://www.tiingo.com/documentation/boats). You can find out about the full product offering on the [Product - BOATS](https://www.tiingo.com/products/boats-blue-ocean-ats-real-time-overnight-stock-prices) page.

## 3.5.2 Top-of-Book & Last Trade

With Tiingo's Websocket/Firehose BOATS API, you receive every validated top-of-book and last-trade update from the BOATS venue feed.

Use thresholdLevel **3** (the only public BOATS channel today) to subscribe to the full venue firehose.

For the BOATS Websocket API:

- A "thresholdLevel" of **3** means you will get ALL venue-native Top-of-Book AND Last Trade updates from BOATS.

**To request top-of-book and last trade data, use the following Websocket endpoint.**

```
# Websocket Top-of-Book & Last Trade Endpoint
wss://api.tiingo.com/boats
```

### Request

To subscribe to the Websocket API you must first open a connection then send the following request parameters. Check out the "Examples" tab to see copy & paste examples of how this works.

| FIELD NAME | PARAMETER | JSON FIELD | DATA TYPE | REQUIRED | DESCRIPTION |
| --- | --- | --- | --- | --- | --- |
| Event Name | Websocket | eventName | string | Y | This will be either "subscribe" or "unsubscribe". |
| Authorization | Websocket | authorization | string | Y | This will be your API Auth token. |
| Event Data | Websocket | eventData | object | N | A JSON object passed to the websocket that contains subscribe/unsubscribe parameters. This lets you set the "thresholdLevel". See the table below for specifics. |

The following parameters are what can be passed in the "eventData" field.

#### "eventData" Request Parameters

| FIELD NAME | PARAMETER | JSON FIELD | DATA TYPE | REQUIRED | DESCRIPTION |
| --- | --- | --- | --- | --- | --- |
| Threshold Level | Websocket | thresholdLevel | string | Y | A threshold level. This is a filter that determines how much data you get from the BOATS feed.<br>3 - All venue-native BOATS top-of-book and last-trade updates (full firehose). Requires the BOATS Real-time entitlement. |

### Response

The BOATS websocket returns meta information about the websocket update message along with the raw data related to that update message.

Check out the table below to see the top-level fields returned from the Websocket BOATS API.

| FIELD NAME | JSON FIELD | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Service code | service | string | An identifier telling you this is BOATS data. The value returned by this will always be "boats". |
| Message Type | messageType | char | A value telling you what kind of data packet this is from our BOATS feed. Will always return "A" meaning new price quotes. |
| BOATS TOP/Price Data | data | array | An array containing the BOATS data. Quote ("Q") and Trade ("T") / Break ("B") updates use different compact layouts — see the tables below. This is a venue-native firehose shape, not the longer IEX /equities array. |

To see what fields are returned in the "data" field, please see the table below.

#### Top-of-Book (Quote) Update Messages

Quote ("Q") updates return a compact 9-element `data` array with the venue top-of-book only (bid/mid/ask). Halt, after-hours, and trade-slot placeholders from the older IEX-style layout are not included on this feed.

| FIELD NAME | ARRAY INDEX | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Update Message Type | 0 | char | Communicates what type of price update this is. Always "Q" for top-of-book quote updates. |
| Date | 1 | datetime | A string representing the datetime this quote came in. This is the timestamp reported by BOATS in JSON ISO Format. |
| Nanoseconds | 2 | int64 | An integer representing the number of nanoseconds since POSIX (Epoch) time UTC. |
| Ticker | 3 | string | Ticker related to the asset. |
| Bid Size | 4 | int32 | The number of shares at the bid price. |
| Bid Price | 5 | float | The current highest bid price. |
| Mid Price | 6 | float | The mid price when both bidPrice and askPrice are positive. In mathematical terms:<br>mid = (bidPrice + askPrice)/2.0<br>This value is calculated by Tiingo and not provided by BOATS.<br>Null when either side is missing or non-positive (for example a one-sided book). |
| Ask Price | 7 | float | The current lowest ask price. |
| Ask Size | 8 | int32 | The number of shares at the ask price. |

#### Trade & Break Update Messages

Trade ("T") and Break ("B") updates return a compact 10-element `data` array: last price/size plus the four raw MEMOIR Sale Condition characters at indexes 6–9 (so values like "H" vs "X" stay distinct).

| FIELD NAME | ARRAY INDEX | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Update Message Type | 0 | char | Communicates what type of price update this is. "T" for last trade, "B" for trade break. |
| Date | 1 | datetime | A string representing the datetime this trade came in. This is the timestamp reported by BOATS in JSON ISO Format. |
| Nanoseconds | 2 | int64 | An integer representing the number of nanoseconds since POSIX (Epoch) time UTC. |
| Ticker | 3 | string | Ticker related to the asset. |
| Last Price | 4 | float | The price at which the last trade was executed. |
| Last Size | 5 | int32 | The amount of shares (volume) traded at the last price. |
| Sale Condition 1 | 6 | char | The first MEMOIR Last Sale sale-condition character from BOATS. Whitespace-padded venue slots are returned as an empty string "". Commonly "@" for a regular sale. |
| Sale Condition 2 | 7 | char | The second MEMOIR Last Sale sale-condition character from BOATS. "F" indicates an Intermarket Sweep Order (ISO) / trade-through exemption. Unused positions are "". |
| Sale Condition 3 | 8 | char | The third MEMOIR Last Sale sale-condition character from BOATS. "T" indicates a Form T / extended-hours trade. Unused positions are "". |
| Sale Condition 4 | 9 | char | The fourth MEMOIR Last Sale sale-condition character from BOATS. Notable values include "I" (odd lot), "H" (price variation), and "X" (cross / periodic auction). "H" and "X" are kept distinct so clients do not lose that MEMOIR distinction. Unused positions are "". |

### Example

```python
from websocket import create_connection
import simplejson as json
ws = create_connection("wss://api.tiingo.com/boats")

subscribe = {
        'eventName':'subscribe',
        'authorization':'<TOKEN>',
        'eventData': {
            'thresholdLevel': 3
    }
}

ws.send(json.dumps(subscribe))
while True:
    print(ws.recv())
```

**Response:**

```json
{
    "messageType":"A",
    "service":"boats",
    "data":[
        "Q",
        "2019-01-30T13:33:45.383129126-05:00",
        1548873225383129126,
        "vym",
        100,
        81.58,
        81.585,
        81.59,
        100
    ]
}
{
    "messageType":"A",
    "service":"boats",
    "data":[
        "T",
        "2019-01-30T13:33:45.594808294-05:00",
        1548873225594808294,
        "wes",
        50.285,
        200,
        "@",
        "F",
        "",
        "X"
    ]
}
```

The below example shows how you can subscribe to a few tickers and then update the subscription to add or remove new tickers to the list.

First you can create a subscription with the "tickers" parameter. If you pass "*" as a ticker, it will mean you will get data for ALL tickers.

```python
from websocket import create_connection
import simplejson as json
ws = create_connection("wss://api.tiingo.com/boats")

subscribe = {
        'eventName':'subscribe',
        'authorization':'<TOKEN>',
        'eventData': {
            'thresholdLevel': 3,
            'tickers': ['spy', 'uso']
    }
}

ws.send(json.dumps(subscribe))
while True:
    print(ws.recv())
```

**Response:**

```json
{
    "messageType":"I",
    "data": {
        "subscriptionId":13706
    },
    "response": {
        "code":200,
        "message":"Success"
    }
}
{
    "messageType":"H",
    "response": {
        "code":200,
        "message":"HeartBeat"
    }
}
{
    "messageType":"A",
    "service":"boats"
    "data":[
        "Q",
        "2019-02-14T12:17:19.342553795-05:00",
        1550164639342553795,
        "spy",
        100,
        274.58,
        274.595,
        274.61,
        100
    ],
}
{
    "messageType":"A",
    "service":"boats"
    "data":[
        "Q",
        "2019-02-14T12:17:20.105597077-05:00",
        1550164640105597077,
        "uso",
        20000,
        11.39,
        11.395,
        11.4,
        3000
    ]
}
```

Use the **subscriptionId** from the prior response to update the subscription. Specify the **subscriptionId** in the **eventData** parameter.

This command should be run as a second process. This means you can edit the first concurrent subscription without having to stop the flow of data packets.

The below example removes "spy" as one of the subscribed tickers. You will be returned a list of currently subscribed tickers, in the response, if the action is successful.

```python
from websocket import create_connection
import simplejson as json
ws = create_connection("wss://api.tiingo.com/boats")

subscribe = {
        'eventName':'unsubscribe',
        'authorization':'<TOKEN>',
        'eventData': {
            'subscriptionId': 13706,
            'tickers': ['spy']
    }
}

ws.send(json.dumps(subscribe))
print(ws.recv())
```

**Response:**

```json
{
    "data": {
        "tickers":[
            "uso"
        ],
        "thresholdLevel":"3"
    },
    "messageType":"I",
    "response": {
        "code":200,
        "message":"Success"
    }
}
```

The below examples add all tickers ("*") as one of the subscribed tickers. You will be returned a list of currently subscribed tickers, in the response, if the action is successful.

```python
from websocket import create_connection
import simplejson as json
ws = create_connection("wss://api.tiingo.com/boats")

subscribe = {
        'eventName':'subscribe',
        'authorization':'<TOKEN>',
        'eventData': {
            'subscriptionId': 13706,
            'tickers': ['*']
    }
}

ws.send(json.dumps(subscribe))
print(ws.recv())
```

**Response:**

```json
{
    "data": {
        "tickers":[
            "*",
            "uso"
        ],
        "thresholdLevel":"3"
    },
    "messageType":"I",
    "response": {
        "code":200,
        "message":"Success"
    }
}
```

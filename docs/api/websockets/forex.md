# Real-time Forex WebSocket API Documentation

> 출처: https://www.tiingo.com/documentation/websockets/forex · 생성: 2026-09-04 (tools/gendocs)

## Endpoints

**WebSocket Endpoints**

```
# Websocket Forex API Top-of-Book Endpoint
wss://api.tiingo.com/fx
```

## 3.2.1 Overview

Tiingo provides updates via websocket every time the Top-of-book (best bid/offer price) change.

We obtain our data through raw binary feeds we receive via feeds that cannot to tier-1 banks and FX dark pools. We then send the data straight to you after minor processing (even before we update our databases).

To further minimize latency, we use bare metal machines for our cloud infrastructure which are located about 15 miles from the NY5 data center.

With Tiingo Free, Power, Commercial, and Redistribution plans all come with access to the firehose. Please note the firehose exposes a very high amount of data, in some cases to the microsecond resolution. Please build your systems cautiously and to scale, otherwise you may use our REST API which leverages Tiingo's infrastructure for this purpose.

You can find out about the full product offering on the [Product - Forex](https://www.tiingo.com/products/forex-api) page.

## 3.2.2 Top-of-Book

With Tiingo's Websocket/Firehose Forex API, you can gain access to all data we receive via our feeds.

To control how much data you would like to receive, read about the "thresholdLevel" request parameter below. A higher "thresholdLevel" means you will get less updates, which could potentially be more relevant.

For the Forex Websocket API:

- A "thresholdLevel" of 5 means you will get ALL Top-of-Book updates.

**To request top-of-book and last trade data, use the following Websocket endpoint.**

```
# Websocket Top-of-Book Endpoint
wss://api.tiingo.com/fx
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
| Subscription ID | Websocket | subscriptionId | string | N | This is used to modify existing subscriptions. Leave this blank for new connections. Can use this parameter to subscribe/unsubscribe from additional tickers. |
| Threshold Level | Websocket | thresholdLevel | string | N | A threshold level. This is a filter that determines how much data you get from the FX feed.<br>7 - A top-of-book update that is due to a change in either the bid/ask price or size. |

### Response

The Forex websocket returns meta information about the websocket update message along with the raw data related to that update message.

Check out the table below to see the top-level fields returned from the Websocket FX API.

| FIELD NAME | JSON FIELD | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Service code | service | string | An identifier telling you this is FX data. The value returned by this will always be "fx". |
| Message Type | messageType | char | A value telling you what kind of data packet this is from our FX feed. Will always return "A" meaning new price quotes. |
| FX TOP/Price Data | data | array | An array containing the FX data. See the tables below to see the fields returned in the array depending if the message is a "Trade" or "TOP" Update message. |

To see what fields are returned in the "data" field, please see the table below.

#### Top-of-Book (Quote) Update Messages

| FIELD NAME | ARRAY INDEX | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Update Message Type | 0 | char | Communicates what type of price update this is. Will always be "Q" for top-of-book update message. |
| Ticker | 1 | string | Ticker related to the asset. |
| Date | 2 | datetime | A string representing the datetime this quote or trade came in. This is reported in JSON ISO Format. |
| Bid Size | 3 | float | The number shares at the bid price. |
| Bid Price | 4 | float | The current highest bid price. |
| Mid Price | 5 | float | The mid price of the current timestamp when both "bidPrice" and "askPrice" are not-null. In mathematical terms:<br>mid = (bidPrice + askPrice)/2.0 |
| Ask Size | 7 | float | The number of units at the ask price. |
| Ask Price | 6 | float | The current lowest ask price. |

### Example

```python
from websocket import create_connection
import simplejson as json
ws = create_connection("wss://api.tiingo.com/fx")

subscribe = {
        'eventName':'subscribe',
        'authorization':'<TOKEN>',
        'eventData': {
            'thresholdLevel': 5
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
    "service":"fx",
    "data":[
        "Q",
        "eurnok",
        "2019-07-05T15:49:15.157000+00:00",
        5000000.0,
        9.6764,
        9.678135,
        5000000.0,
        9.67987
    ]
}
{
    "messageType":"A",
    "service":"fx",
    "data":[
        "Q",
        "gbpaud",
        "2019-07-05T15:49:15.236000+00:00",
        1000000.0,
        1.79457,
        1.79477,
        5000000.0,
        1.79497
    ]
}
```

The below example shows how you can subscribe to a few tickers and then update the subscription to add or remove new tickers to the list.

First you can create a subscription with the "tickers" parameter. If you pass "*" as a ticker, it will mean you will get data for ALL tickers.

```python
from websocket import create_connection
import simplejson as json
ws = create_connection("wss://api.tiingo.com/fx")

subscribe = {
        'eventName':'subscribe',
        'authorization':'<TOKEN>',
        'eventData': {
            'thresholdLevel': 5,
            'tickers': ['audusd', 'eurusd']
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
    "service":"fx",
    "data":[
        "Q",
        "eurusd",
        "2023-06-14T14:25:50.432000+00:00",
        1000000.0,
        1.0847,
        1.084745,
        1000000.0,
        1.08479
    ]
}
{
    "messageType":"A",
    "service":"fx",
    "data":[
        "Q",
        "audusd",
        "2023-06-14T14:25:50.557000+00:00",
        1000000.0,
        0.68298,
        0.68303,
        1000000.0,
        0.68308
    ]
}
```

Use the **subscriptionId** from the prior response to update the subscription. Specify the **subscriptionId** in the **eventData** parameter.

This command should be run as a second process. This means you can edit the first concurrent subscription without having to stop the flow of data packets.

The below example removes "audusd" as one of the subscribed tickers. You will be returned a list of currently subscribed tickers, in the response, if the action is successful.

```python
from websocket import create_connection
import simplejson as json
ws = create_connection("wss://api.tiingo.com/fx")

subscribe = {
        'eventName':'unsubscribe',
        'authorization':'<TOKEN>',
        'eventData': {
            'subscriptionId': 13706,
            'tickers': ['audusd']
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
            "eurusd"
        ],
        "thresholdLevel":"5"
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
ws = create_connection("wss://api.tiingo.com/fx")

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
            "eurusd"
        ],
        "thresholdLevel":"0"
    },
    "messageType":"I",
    "response": {
        "code":200,
        "message":"Success"
    }
}
```

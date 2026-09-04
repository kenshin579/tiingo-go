# Real-time Consolidated Equity WebSocket API Documentation

> 출처: https://www.tiingo.com/documentation/websockets/equity-realtime-stock-data · 생성: 2026-09-04 (tools/gendocs)

## Endpoints

**WebSocket Endpoints**

The Equity Realtime Websocket Endpoints are currently in beta. For production use cases, we recommend the [IEX websocket endpoints](https://www.tiingo.com/documentation/websockets/iex) - which these endpoints expand upon.

```
# Websocket Consolidated Equity Reference Price & Liquidity Risk Metric Endpoint
wss://api.tiingo.com/equity/intraday
```

## 3.3.1 Overview

Tiingo provides consolidated equity websocket updates from Tiingo's connections into multiple exchanges, ATS, and OTC venues. The feed combines multiple US equity market data sources into one stream and exposes two derived products: Reference Prices and Lquidity Bid/Ask metrics. These are accessible via two different `thresholdLevel` subscriptions.

- **thresholdLevel 6** publishes reference price updates when Tiingo detects a meaningful consolidated reference price change.
- **thresholdLevel 4** publishes liquidity spread (bid/ask) updates, including `lqSpread` and the related `lq*` bid/ask fields (price and sizes).

The liquidity risk metric is a statistical estimate of an asset's expected bid/ask spread and liquidity at top-of-book. Tiingo gathers spread and liquidity observations from multiple venues - including exchanges, ATS, and OTC venues - and applies a proprietary methodology to produce an estimated liquidity spread and liquidity bid/ask profile for each security.

For REST snapshots and historical intraday bars, see the [Equity Realtime REST documentation](https://www.tiingo.com/documentation/equity-realtime-stock-data).

## 3.3.2 Reference Price (thresholdLevel 6)

Use `thresholdLevel` 6 to receive consolidated reference price updates. Each message contains a timestamp, ticker, and reference price. This reference price corresponds to `tngoLast` in the REST API and is calculated from validated consolidated quote mids or trade prints. These values are used to create OHLC bars in the REST endpoints.

- A `thresholdLevel` of **6** means you receive reference price updates when Tiingo detects a meaningful consolidated reference price change.

**To request consolidated reference price data, use the following Websocket endpoint.**

```
# Websocket Consolidated Equity Reference Price Endpoint
wss://api.tiingo.com/equity/intraday
```

### Request

To subscribe to the Websocket API you must first open a connection then send the following request parameters. Check out the "Examples" tab to see copy & paste examples of how this works.

| FIELD NAME | PARAMETER | JSON FIELD | DATA TYPE | REQUIRED | DESCRIPTION |
| --- | --- | --- | --- | --- | --- |
| Event Name | Websocket | eventName | string | Y | This will be either "subscribe" or "unsubscribe". |
| Authorization | Websocket | authorization | string | Y | This will be your API Auth token. |
| Event Data | Websocket | eventData | object | N | A JSON object passed to the websocket that contains subscribe/unsubscribe parameters. This lets you set the "thresholdLevel". See the table below for specifics. |

The following parameters are what can be passed in the `eventData` field.

#### "eventData" Request Parameters

| FIELD NAME | PARAMETER | JSON FIELD | DATA TYPE | REQUIRED | DESCRIPTION |
| --- | --- | --- | --- | --- | --- |
| Threshold Level | Websocket | thresholdLevel | string | Y | A threshold level. This is a filter that determines how much data you get from the consolidated equity feed.<br>6 - Receive consolidated reference price updates when Tiingo detects a meaningful reference price change. |

### Response

The websocket returns meta information about the update message along with the compact reference price array.

| FIELD NAME | JSON FIELD | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Service code | service | string | An identifier telling you this is consolidated equity data. The value returned by this will always be "cons". |
| Message Type | messageType | char | A value telling you what kind of data packet this is. Will always return "A" meaning new price quotes. |
| Consolidated Reference Price Data | data | array | An array containing the consolidated reference price data. See the table below for the fields returned in the array. |

To see what fields are returned in the `data` field, please see the table below.

#### Reference Price Update Messages

| FIELD NAME | ARRAY INDEX | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Date | 0 | datetime | A string representing the datetime this reference price update was published, in JSON ISO format. |
| Ticker | 1 | string | Ticker related to the asset. |
| Reference Price | 2 | float | Tiingo's consolidated reference price for the asset. This is calculated from validated consolidated quote mids or trade prints across enabled equity venues and corresponds to tngoLast in the REST API. |

### Example

```python
from websocket import create_connection
import simplejson as json
ws = create_connection("wss://api.tiingo.com/equity/intraday")

subscribe = {
        'eventName':'subscribe',
        'authorization':'<TOKEN>',
        'eventData': {
            'thresholdLevel': 6,
            'tickers': ['*']
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
    "service":"cons",
    "data":[
        "2019-01-30T13:33:45.383129126-05:00",
        "aapl",
        162.33
    ]
}
{
    "messageType":"A",
    "service":"cons",
    "data":[
        "2019-01-30T13:33:45.594808294-05:00",
        "spy",
        274.595
    ]
}
```

## 3.3.3 Liquidity Risk Metric (thresholdLevel 4)

Use `thresholdLevel` 4 to receive liquidity spread updates. Each message contains a timestamp, ticker, liquidity spread, liquidity bid size, liquidity bid price, reference price, liquidity ask price, and liquidity ask size.

The payload shape is `[timestamp, ticker, lqSpread, lqBidSize, lqBidPrice, referencePrice, lqAskPrice, lqAskSize]`. `lqSpread` is a relative decimal (i.e. where 0.04 means 4%). The remaining liquidity fields correspond to the `lq*` fields exposed on the consolidated equity REST endpoint, e.g. lqBidPrice and lqAskPrice. These are expanded upon below. These metrics are intended to be used as valuation and risk metrics to help value assets in real-time.

- A `thresholdLevel` of **4** means you receive liquidity risk metric updates when Tiingo republishes the statistical liquidity estimate for a security.

**To request liquidity risk metric data, use the following Websocket endpoint.**

```
# Websocket Consolidated Equity Liquidity Risk Metric Endpoint
wss://api.tiingo.com/equity/intraday
```

### Request

To subscribe to the Websocket API you must first open a connection then send the following request parameters. Check out the "Examples" tab to see copy & paste examples of how this works.

| FIELD NAME | PARAMETER | JSON FIELD | DATA TYPE | REQUIRED | DESCRIPTION |
| --- | --- | --- | --- | --- | --- |
| Event Name | Websocket | eventName | string | Y | This will be either "subscribe" or "unsubscribe". |
| Authorization | Websocket | authorization | string | Y | This will be your API Auth token. |
| Event Data | Websocket | eventData | object | N | A JSON object passed to the websocket that contains subscribe/unsubscribe parameters. This lets you set the "thresholdLevel". See the table below for specifics. |

The following parameters are what can be passed in the `eventData` field.

#### "eventData" Request Parameters

| FIELD NAME | PARAMETER | JSON FIELD | DATA TYPE | REQUIRED | DESCRIPTION |
| --- | --- | --- | --- | --- | --- |
| Threshold Level | Websocket | thresholdLevel | string | Y | A threshold level. This is a filter that determines how much data you get from the consolidated equity feed.<br>4 - Receive liquidity risk metric updates — statistical estimates of spread and liquidity characteristics derived from aggregated historical observations. |

### Response

The websocket returns meta information about the update message along with the compact liquidity risk metric array.

| FIELD NAME | JSON FIELD | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Service code | service | string | An identifier telling you this is consolidated equity data. The value returned by this will always be "cons". |
| Message Type | messageType | char | A value telling you what kind of data packet this is. Will always return "A" meaning a new liquidity risk metric update. |
| Consolidated Liquidity Risk Metric Data | data | array | An array containing the statistical liquidity risk metric data. See the table below for the fields returned in the array. |

To see what fields are returned in the `data` field, please see the table below.

#### Liquidity Risk Metric Update Messages

| FIELD NAME | ARRAY INDEX | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Date | 0 | datetime | A string representing the datetime this liquidity risk metric update was published, in JSON ISO format. |
| Ticker | 1 | string | Ticker related to the asset. |
| Liquidity Spread | 2 | float | The relative spread width component of the liquidity risk metric, expressed as a decimal where 0.04 means 4%. Corresponds to lqSpread in the REST API. |
| Liquidity Bid Size | 3 | int32 | The bid size component of the liquidity risk metric in shares. Corresponds to lqBidSize in the REST API. |
| Liquidity Bid Price | 4 | float | The bid price component of the liquidity risk metric. Corresponds to lqBidPrice in the REST API. |
| Reference Price | 5 | float | The reference price used as the center of the liquidity risk metric. This is normally the threshold-level 6 reference price (tngoLast) when available. |
| Liquidity Ask Price | 6 | float | The ask price component of the liquidity risk metric. Corresponds to lqAskPrice in the REST API. |
| Liquidity Ask Size | 7 | int32 | The ask size component of the liquidity risk metric in shares. Corresponds to lqAskSize in the REST API. |

### Example

```python
from websocket import create_connection
import simplejson as json
ws = create_connection("wss://api.tiingo.com/equity/intraday")

subscribe = {
        'eventName':'subscribe',
        'authorization':'<TOKEN>',
        'eventData': {
            'thresholdLevel': 4,
            'tickers': ['aapl', 'spy']
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
    "service":"cons",
    "data":[
        "2019-01-30T13:33:45.383129126-05:00",
        "aapl",
        0.0011,
        100,
        162.30,
        162.33,
        162.36,
        100
    ]
}
{
    "messageType":"A",
    "service":"cons",
    "data":[
        "2019-01-30T13:33:45.594808294-05:00",
        "spy",
        0.0008,
        100,
        274.57,
        274.595,
        274.62,
        100
    ]
}
```

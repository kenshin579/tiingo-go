# Equity Realtime Stock Data API Documentation

> 출처: https://www.tiingo.com/documentation/equity-realtime-stock-data · 생성: 2026-09-04 (tools/gendocs)

## Endpoints

**REST Endpoints**

The Equity Realtime API Endpoints are currently in beta. For production use cases, we recommend the [IEX endpoints](https://www.tiingo.com/documentation/iex) - which these endpoints expand upon.

```
# To request reference price and liquidity metrics for all tickers, use the following REST endpoint
https://api.tiingo.com/tiingo/equity/intraday

# To request reference price and liquidity metrics for specific tickers, use the following REST endpoint
https://api.tiingo.com/tiingo/equity/intraday/<ticker>

# Historical Intraday Prices
https://api.tiingo.com/tiingo/equity/intraday/<ticker>/prices?startDate=2019-01-02&resampleFreq=5min
```

## 2.5.1 Overview

This endpoint is an alternative to the IEX Endpoints and creates datasets from multiple Equity venues (exchanges, ATS, and OTC venues). Tiingo is the first company to make these exchange-compliant derived metrics and offer them for free.

This changes the game on exchange liquidity and the products we can offer to provide better reference pricing for risk and valuation metrics for your portfolio valuations, risk assessments, trading, and app/AI development.

The data is sourced from Tiingo's consolidated equity pipeline and exposed through `/tiingo/equity/intraday`. Historical bars use consolidated intraday data where available and preserve the familiar start date, end date, resampling, after-hours, and force-fill controls.

Benefits of Tiingo Equity Realtime

- Realtime consolidated equity snapshots for supported US equities and ETFs.
- Market hours from 4am ET to 8pm ET.
- Tiingo reference price, liquidity reference price, liquidity spread and bid/ask metrics, and day-level OHLC fields.
- Historical intraday OHLCV bars with configurable minute and hourly resampling.
- Optional after-hours data and force-filled bars for charting workflows.
- Data is served via REST endpoints and available via [WebSockets](https://www.tiingo.com/documentation/websockets/equity-realtime-stock-data) as well.

## 2.5.2 Current Reference Price & Liquidity Snapshot

**To request reference price and liquidity metrics for a stock, use the following REST endpoints.**

```
# To request reference price and liquidity metrics for all tickers, use the following REST endpoint
https://api.tiingo.com/tiingo/equity/intraday

# To request reference price and liquidity metrics for specific tickers, use the following REST endpoint
https://api.tiingo.com/tiingo/equity/intraday/<ticker>
```

### Request

| FIELD NAME | PARAMETER | JSON FIELD | DATA TYPE | REQUIRED | DESCRIPTION |
| --- | --- | --- | --- | --- | --- |
| Ticker | URL or GET | tickers | string | N | Ticker related to the asset. |
| Response Format | GET | format | string | N | Sets the response format of the returned data. Acceptable values are "csv" and "json". Defaults to JSON. |

### Response

Fields may be null when the underlying consolidated feed has not published that value.

| FIELD NAME | JSON FIELD | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Ticker | ticker | string | Ticker related to the asset. |
| Timestamp | timestamp | datetime | The timestamp the data was last refresh on. |
| Tiingo Last | tngoLast | float | Tiingo Last is either the last price or mid price. The mid price is only used if our algo determines it is a good proxy for the last price. So if the spread is considered wide by our algo, we do not use it. Also, after the official exchange print comes in, this value changes to that value. This value is calculated by Tiingo from the consolidated equity feed. |
| Liquidity Reference Price | lqRefPrice | float | The same as tngoLast - mirrored for convenience/consistency as it's better-named for its role in the lq spread metrics. |
| Previous Close | prevClose | float | Previous day's closing price of the security. This can come from the supported consolidated equity market data sources. |
| Open | open | float | The opening price of the asset on the current day This value is calculated by Tiingo from the consolidated equity feed. |
| High | high | float | The high price of the asset on the current day This value is calculated by Tiingo from the consolidated equity feed. |
| Low | low | float | The low price of the asset on the current day This value is calculated by Tiingo from the consolidated equity feed. |
| Volume | volume | int64 | Volume will be consolidated intraday volume throughout the day. Once the official closing price comes in, volume may reflect the full official trading day. This field is available for convenience. |
| Liquidity Spread | lqSpread | float | The relative lqBid/Ask spread component of the liquidity risk metric, expressed as a decimal (e.g. 0.04 means 4%). Corresponds to lqSpread in the thresholdLevel 4 websocket liquidity risk metric. |
| Liquidity Bid Price | lqBidPrice | float | The bid price component of the liquidity risk metric. Corresponds to lqBidPrice in the thresholdLevel 4 websocket liquidity risk metric. |
| Liquidity Bid Size | lqBidSize | int64 | The bid size component of the liquidity risk metric in shares. Corresponds to lqBidSize in the thresholdLevel 4 websocket liquidity risk metric. |
| Liquidity Ask Price | lqAskPrice | float | The ask price component of the liquidity risk metric. Corresponds to lqAskPrice in the thresholdLevel 4 websocket liquidity risk metric. |
| Liquidity Ask Size | lqAskSize | int64 | The ask size component of the liquidity risk metric in shares. Corresponds to lqAskSize in the thresholdLevel 4 websocket liquidity risk metric. |

### Example

```python
import requests
headers = {
    'Content-Type': 'application/json'
}
requestResponse = requests.get("https://api.tiingo.com/tiingo/equity/intraday/?tickers=aapl,spy&token=<TOKEN>", headers=headers)
print(requestResponse.json())
```

**Response:**

```json
[
    {
        "ticker":"AAPL",
        "timestamp":"2019-01-30T10:33:38.186520297-05:00",
        "tngoLast":162.33,
        "lqRefPrice":162.33,
        "prevClose":154.68,
        "open":161.83,
        "high":163.25,
        "low":160.38,
        "volume":0,
        "lqSpread":0.0011,
        "lqBidPrice":162.30,
        "lqBidSize":100,
        "lqAskPrice":162.36,
        "lqAskSize":100
    },
    {
        "ticker":"SPY",
        "timestamp":"2019-01-30T11:12:29.505261845-05:00",
        "tngoLast":265.405,
        "lqRefPrice":265.405,
        "prevClose":263.41,
        "open":264.62,
        "high":265.445,
        "low":264.225,
        "volume":0,
        "lqSpread":0.0008,
        "lqBidPrice":265.39,
        "lqBidSize":500,
        "lqAskPrice":265.42,
        "lqAskSize":100
    }
]
```

## 2.5.3 Historical Intraday Prices Endpoint

**To request historical intraday prices for a stock, use the following REST endpoint.**

```
# Historical Intraday Prices
https://api.tiingo.com/tiingo/equity/intraday/<ticker>/prices?startDate=2019-01-02&resampleFreq=5min
```

### Request

| FIELD NAME | PARAMETER | JSON FIELD | DATA TYPE | REQUIRED | DESCRIPTION |
| --- | --- | --- | --- | --- | --- |
| Ticker | URL | N/A | string | Y | Ticker related to the asset. |
| Start Date | GET | startDate | date | N | If startDate or endDate is not null, historical data will be queried. This filter limits metrics to on or after the startDate (>=). Parameter must be in YYYY-MM-DD format. |
| End Date | GET | endDate | date | N | If startDate or endDate is not null, historical data will be queried. This filter limits metrics to on or before the endDate (<=). Parameter must be in YYYY-MM-DD format. |
| Resample Freq | GET | resampleFreq | string | N | This allows you to set the frequency in which you want data resampled. For example "1hour" would return the data where OHLC is calculated on an hourly schedule. The minimum value is "1min". Both units in minutes (min) and hours (hour) are accepted. Format is # + (min/hour); e.g. "15min" or "4hour". If no value is provided, defaults to 5min. |
| After Hours | GET | afterHours | boolean | N | If set to true, includes pre and post market data if available. |
| Force Fill | GET | forceFill | boolean | N | Some tickers do not have a trade/quote update for a given time period. if forceFill is set to true, then the previous OHLC will be used to fill the current OHLC. |
| Response Format | GET | format | string | N | Sets the response format of the returned data. Acceptable values are "csv" and "json". Defaults to JSON. |

### Response

| FIELD NAME | JSON FIELD | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Date | date | datetime | The date this data pertains to. |
| Open | open | float | The opening price for the asset on the given date. |
| High | high | float | The high price for the asset on the given date. |
| Low | low | float | The low price for the asset on the given date. |
| Close | close | float | The closing price for the asset on the given date. |
| Volume | volume | int64 | The consolidated number of shares traded for the interval. This value will only be exposed if explicitly passed to the "columns" request parameter. E.g. ?columns=open,high,low,close,volume |

### Example

```python
import requests

headers = {
    'Content-Type': 'application/json'
}
requestResponse = requests.get("https://api.tiingo.com/tiingo/equity/intraday/aapl/prices?startDate=2019-01-02&resampleFreq=5min&columns=open,high,low,close,volume&token=<TOKEN>", headers=headers)
print(requestResponse.json())
```

**Response:**

```json
[
    {
        "date":"2019-01-02T14:30:00.000Z",
        "open":154.74,
        "high":155.52,
        "low":154.58,
        "close":154.76,
        "volume":16102
    },
    {
        "date":"2019-01-02T14:35:00.000Z",
        "open":154.8,
        "high":155.0,
        "low":154.31,
        "close":154.645,
        "volume":19127
    }
    ...
]
```

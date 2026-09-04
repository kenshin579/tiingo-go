# Stock, ETF, and Mutual Fund Split API Documentation

> 출처: https://www.tiingo.com/documentation/corporate-actions/splits · 생성: 2026-09-04 (tools/gendocs)

## Stock, ETF, and Mutual Funds Splits API Documentation

**REST Endpoints**

The new Corporate Action endpoints, such as the splits endpoints below, are currently available to Beta-enabled customers and enterprise customers as an early release product. If you would like access to the data below, please E-mail [support@tiingo.com](mailto:support@tiingo.com)

```
# Split data batch endpoint
https://api.tiingo.com/tiingo/corporate-actions/splits

# Split data endpoint for specific ticker
https://api.tiingo.com/tiingo/corporate-actions/<ticker>/splits
```

## 2.11.1 Overview

The Tiingo API Splits API endpoint provides detailed split data, for both historical and future split data. This endpoint remains newly exposed in beta and is included for all customers with the End-of-Day endpoint entitlement.

Splits are available for current and future dates, and data is updated as new corporate communications get processed. The tickers covered are the same found via the [End-of-Day price data](https://www.tiingo.com/documentation/end-of-day) endpoints. This mean you can get past and future split data for stocks, ETFs, and mutual funds.

You can find out about the full product offering on the [Product - Stock, ETF, & Mutual Fund Split API](https://www.tiingo.com/products/corporate-actions/stock-etf-mutual-fund-splits-api) page.

## 2.11.2 Batch Split Data

Use this endpoint to get past, present, and future splits. This endpoint will return detailed split data for a given stock, ETF, or Mutual Fund. You will also notice split status - this can either be active ("a") or cancelled ("c").

**To request distribution data for a stock, use the following REST endpoints.**

```
# Latest Split Data with an Ex-Date of the current day
https://api.tiingo.com/tiingo/corporate-actions/splits

# Latest Split Data with an ex-date explicitly specified (future date or historical date)
https://api.tiingo.com/tiingo/corporate-actions/splits?exDate=2023-08-25
```

### Request

| FIELD NAME | PARAMETER | JSON FIELD | DATA TYPE | REQUIRED | DESCRIPTION |
| --- | --- | --- | --- | --- | --- |
| Ex-Date | GET | exDate | date | N | This filter limits distributions that have an ex-date on the date passed. Parameter must be in YYYY-MM-DD format. |
| Response Format | GET | format | string | N | Sets the response format of the returned data. Acceptable values are "csv" and "json". Defaults to JSON. |
| Columns | GET | columns | string[] | N | Allows you to specify which columns you would like returned from the output. Pass an array of strings of column names to get only this columns back. |

### Response

| FIELD NAME | JSON FIELD | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| PermaTicker | permaTicker | string | The Tiingo permaticker. |
| Ticker | ticker | string | Ticker related to the asset. |
| Ex-Date | exDate | datetime | The ex-Date of the split. In the Tiingo EOD Endpoints, this is the date where "splitFactor" will not be 1.0. This is also the date used for split adjustments. |
| Split From | splitFrom | float | The prior split ratio. |
| Split To | splitTo | float | The new split ratio, i.e. how many shares of "splitTo" are given for each share of "splitFrom". |
| Split Factor | splitFactor | float | The ratio of splitTo from splitFrom. In other words:<br>splitFactor = splitTo/splitFrom<br>This ratio is helpful in calculating split price adjustments. |
| Split Status | splitStatus | string | A code representing the status of split<br>a: Active<br>c: Cancelled |

### Example

```python
import requests
headers = {
    'Content-Type': 'application/json'
}
requestResponse = requests.get("https://api.tiingo.com/tiingo/corporate-actions/splits?exDate=2023-9-28&token=<TOKEN>", headers=headers)
print(requestResponse.json())
```

**Response:**

```json
[
    {
        "permaTicker":"US000000060663",
        "ticker":"dinrf",
        "exDate":"2023-09-28T00:00:00.000Z",
        "splitFrom":1.0,
        "splitTo":2.0,
        "splitFactor":2.0
    }
    ...
]
```

## 2.11.3 Ticker-specific Split Data

This endpoint is similar to the batch endpoint, but allows you to specify a specific ticker to limit the query and also provide historical split timeseries data for a ticker as well. Stocks, ETFs, and Mutual Funds are supported.

**To request distribution data for a stock, use the following REST endpoints.**

```
# Latest Split Data with an Ex-Date of the current day
https://api.tiingo.com/tiingo/corporate-actions/cvs/splits

# Latest Split Data with an ex-date explicitly specified (future date or historical date)
https://api.tiingo.com/tiingo/corporate-actions/cvs/splits?startExDate=2002-08-25
```

### Request

| FIELD NAME | PARAMETER | JSON FIELD | DATA TYPE | REQUIRED | DESCRIPTION |
| --- | --- | --- | --- | --- | --- |
| Ticker | URL | N/A | string | Y | Ticker related to the asset you would like distribution data for. |
| Start Ex-Date | GET | startExDate | date | N | This filter limits distributions that have an ex-date >= on the date passed. Parameter must be in YYYY-MM-DD format. |
| End Ex-Date | GET | endExDate | date | N | This filter limits distributions that have an ex-date <= on the date passed. Parameter must be in YYYY-MM-DD format. |
| Response Format | GET | format | string | N | Sets the response format of the returned data. Acceptable values are "csv" and "json". Defaults to JSON. |
| Columns | GET | columns | string[] | N | Allows you to specify which columns you would like returned from the output. Pass an array of strings of column names to get only this columns back. |

### Response

| FIELD NAME | JSON FIELD | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| PermaTicker | permaTicker | string | The Tiingo permaticker. |
| Ticker | ticker | string | Ticker related to the asset. |
| Ex-Date | exDate | datetime | The ex-Date of the split. In the Tiingo EOD Endpoints, this is the date where "splitFactor" will not be 1.0. This is also the date used for split adjustments. |
| Split From | splitFrom | float | The prior split ratio. |
| Split To | splitTo | float | The new split ratio, i.e. how many shares of "splitTo" are given for each share of "splitFrom". |
| Split Factor | splitFactor | float | The ratio of splitTo from splitFrom. In other words:<br>splitFactor = splitTo/splitFrom<br>This ratio is helpful in calculating split price adjustments. |
| Split Status | splitStatus | string | A code representing the status of split<br>a: Active<br>c: Cancelled |

### Example

```python
import requests
headers = {
    'Content-Type': 'application/json'
}
requestResponse = requests.get("https://api.tiingo.com/tiingo/corporate-actions/dcfc/splits?exDate=2024-03-29&token=<TOKEN>", headers=headers)
print(requestResponse.json())
```

**Response:**

```json
[
    {
        "permaTicker":"US000000102867",
        "ticker":"dcfc",
        "exDate":"2024-04-02T04:00:00.000Z",
        "splitFrom":200.0,
        "splitTo":1.0,
        "splitFactor":0.005
    }
    ...
]
```

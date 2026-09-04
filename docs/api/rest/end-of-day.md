# End-of-Day (EOD) Stock Price API Documentation

> 출처: https://www.tiingo.com/documentation/end-of-day · 생성: 2026-09-04 (tools/gendocs)

## Endpoints

**REST Endpoints**

```
# Meta Data
https://api.tiingo.com/tiingo/daily/<ticker>

# Latest Price
https://api.tiingo.com/tiingo/daily/<ticker>/prices

# Historical Prices
https://api.tiingo.com/tiingo/daily/<ticker>/prices?startDate=2012-1-1&endDate=2016-1-1
```

## 2.1.1 Overview

Tiingo's End-of-Day prices use a proprietary error checking framework to help clean data feeds and also help catch missing corportate actions (splits, dividends, and exchange listing changes). Most US Equity prices are available at 5:30 PM EST, however exchanges may send corrections until 8 PM EST. As we obtain corrections, we update prices throughout the evening.

Mutual Fund NAVs are available after 12 AM EST. The fields "open", "high", "low", "close" will contain the NAV value for the given day.

Both raw prices and adjusted prices are available. The adjustment methodology follows the standard method set forth by "The Center for Research in Security Prices" (CRSP) in this document: [CRSP Calculations](http://www.crsp.com/products/documentation/crsp-calculations). This methodology incorporates both split and dividend adjustments.

You can find out about the full product offering on the [Product - End-of-Day](https://www.tiingo.com/products/end-of-day-stock-price-data) page.

## 2.1.2 End-of-Day Endpoint

**To request price data for a stock, use the following REST endpoints.**

```
# Latest Price Information
https://api.tiingo.com/tiingo/daily/<ticker>/prices

# Historical Price Information
https://api.tiingo.com/tiingo/daily/<ticker>/prices?startDate=2012-1-1&endDate=2016-1-1&format=csv&resampleFreq=monthly
```

### Request

| FIELD NAME | PARAMETER | JSON FIELD | DATA TYPE | REQUIRED | DESCRIPTION |
| --- | --- | --- | --- | --- | --- |
| Ticker | URL | N/A | string | Y | Ticker related to the asset. |
| Start Date | GET | startDate | date | N | If startDate or endDate is not null, historical data will be queried. This filter limits metrics to on or after the startDate (>=). Parameter must be in YYYY-MM-DD format. |
| End Date | GET | endDate | date | N | If startDate or endDate is not null, historical data will be queried. This filter limits metrics to on or before the endDate (<=). Parameter must be in YYYY-MM-DD format. |
| Resample Freq | GET | resampleFreq | string | N | Allows resampled values that allow you to choose the values returned as daily, weekly, monthly, or annually values. Note: ONLY DAILY takes into account holidays. All others use standard business days Acceptable values:<br>daily: Values returned as daily periods, with a holiday calendar.<br>weekly: Values returned as weekly data, with days ending on Friday.<br>monthly: Values returned as monthly data, with days ending on the last standard business day (Mon-Fri) of each month.<br>annually: Values returned as annual data, with days ending on the last standard business day (Mon-Fri) of each year.<br>Note, that if you choose a value in-between the resample period for weekly, monthly, and daily, the start date rolls back to consider the entire period. For example, if you choose to resample weekly, but your "startDate" parameter is set to Wednesday of that week, the startDate will be adjusted to Monday, so the entire week is captured. Another example is if you send a startDate mid-month, we roll back the startDate to the beginning of the month.<br>Similarly, if you provide an endDate, and it's midway through the period, we roll-forward the date to capture the whole period. In the above example, if the end date is set to a wednesday with a weekly resample, the end date is rolled forward to the Friday of that week. |
| Sort | GET | sort | string | N | This field allows you to specify the sort direct and which column to sort by. Prepend "-" if you want descending order, otherwise it will be ascending. E.g. sort=date will sort by date in ascending order. sort=-date will sort by date in descending order. |
| Response Format | GET | format | string | N | Sets the response format of the returned data. Acceptable values are "csv" and "json". Defaults to JSON. |
| Columns | GET | columns | string[] | N | Allows you to specify which columns you would like returned from the output. Pass an array of strings of column names to get only this columns back. |

### Response

| FIELD NAME | JSON FIELD | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Date | date | date | The date this data pertains to. |
| Open | open | float | The opening price for the asset on the given date. |
| High | high | float | The high price for the asset on the given date. |
| Low | low | float | The low price for the asset on the given date. |
| Close | close | float | The closing price for the asset on the given date. |
| Volume | volume | int64 | The number of shares traded for the asset. |
| Adj Open | adjOpen | float | The adjusted opening price for the asset on the given date. |
| Adj High | adjHigh | float | The adjusted high price for the asset on the given date. |
| Adj. Low | adjLow | float | The adjusted low price for the asset on the given date. |
| Adj. Close | adjClose | float | The adjusted closing price for the asset on the given date. |
| Adj. Volume | adjVolume | int64 | The number of shares traded for the asset. |
| Dividend | divCash | float | The dividend paid out on "date" (note that "date" will be the "exDate" for the dividend). |
| Split | splitFactor | float | The factor used to adjust prices when a company splits, reverse splits, or pays a distribution. |

### Example

```python
import requests
headers = {
    'Content-Type': 'application/json'
}
requestResponse = requests.get("https://api.tiingo.com/tiingo/daily/aapl/prices?startDate=2019-01-02&token=<TOKEN>", headers=headers)
print(requestResponse.json())
```

**Response:**

```json
[
    {
        "date":"2019-01-02T00:00:00.000Z",
        "close":157.92,
        "high":158.85,
        "low":154.23,
        "open":154.89,
        "volume":37039737,
        "adjClose":157.92,
        "adjHigh":158.85,
        "adjLow":154.23,
        "adjOpen":154.89,
        "adjVolume":37039737,
        "divCash":0.0,
        "splitFactor":1.0
    },
    {
        "date":"2019-01-03T00:00:00.000Z",
        "close":142.19,
        "high":145.72,
        "low":142.0,
        "open":143.98,
        "volume":91312195,
        "adjClose":142.19,
        "adjHigh":145.72,
        "adjLow":142.0,
        "adjOpen":143.98,
        "adjVolume":91312195,
        "divCash":0.0,
        "splitFactor":1.0
    },
    {
        "date":"2019-01-04T00:00:00.000Z",
        "close":148.26,
        "high":148.5499,
        "low":143.8,
        "open":144.53,
        "volume":58607070,
        "adjClose":148.26,
        "adjHigh":148.5499,
        "adjLow":143.8,
        "adjOpen":144.53,
        "adjVolume":58607070,
        "divCash":0.0,
        "splitFactor":1.0
    },
    {
        "date":"2019-01-07T00:00:00.000Z",
        "close":147.93,
        "high":148.83,
        "low":145.9,
        "open":148.7,
        "volume":54777764,
        "adjClose":147.93,
        "adjHigh":148.83,
        "adjLow":145.9,
        "adjOpen":148.7,
        "adjVolume":54777764,
        "divCash":0.0,
        "splitFactor":1.0
    }
]
```

## 2.1.3 Meta Endpoint

Our meta information comes from a variety of sources, but is used to help communicate details about an asset in our database to our users.

**For a daily list of Tiingo ticker symbols, including currently supported assets and reserved symbols for planned coverage expansion, please see the following file which is updated daily: [supported_tickers.zip](https://apimedia.tiingo.com/docs/tiingo/daily/supported_tickers.zip)**

The supported_tickers.zip file includes current coverage as well as symbol reservations that we have preserved as we expand our data and licensing access for end customers.

**To confirm whether a ticker is presently available through the EOD metadata endpoint, use the following REST endpoint. This endpoint reflects assets currently covered for immediate API access.**

```
# Meta Data
https://api.tiingo.com/tiingo/daily/<ticker>
```

### Request

| FIELD NAME | PARAMETER | JSON FIELD | DATA TYPE | REQUIRED | DESCRIPTION |
| --- | --- | --- | --- | --- | --- |
| Ticker | URL | N/A | string | Y | Ticker related to the asset. |

### Response

| FIELD NAME | JSON FIELD | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Ticker | ticker | string | Ticker related to the asset. |
| Name | name | string | Full-length name of the asset. |
| Exchange Code | exchangeCode | string | An identifier that maps which Exchange this asset is listed on. |
| Description | description | string | Long-form descripton of the asset. |
| Start Date | startDate | date | The earliest date we have price data available for the asset. When null it means no price data available for the given asset. |
| End Date | endDate | date | The latest date we have price data available for the asset. When null it means no price data available for the given asset. |

### Example

```python
import requests

headers = {
    'Content-Type': 'application/json'
}
requestResponse = requests.get("https://api.tiingo.com/tiingo/daily/aapl?token=<TOKEN>", headers=headers)
print(requestResponse.json())
```

**Response:**

```json
{
    "ticker": "AAPL",
    "name": "Apple Inc",
    "exchangeCode": "NASDAQ",
    "startDate": "1980-12-12",
    "endDate": "2019-01-25",
    "description": "Apple Inc. (Apple) designs, manufactures and markets mobile communication and media devices, personal
computers, and portable digital music players, and a variety of related software, services, peripherals, networking
solutions, and third-party digital content and applications. The Company's products and services include iPhone, iPad,
Mac, iPod, Apple TV, a portfolio of consumer and professional software applications, the iOS and OS X operating
systems, iCloud, and a variety of accessory, service and support offerings. The Company also delivers digital content
and applications through the iTunes Store, App StoreSM, iBookstoreSM, and Mac App Store. The Company distributes its
products worldwide through its retail stores, online stores, and direct sales force, as well as through third-party
cellular network carriers, wholesalers, retailers, and value-added resellers. In February 2012, the Company acquired
app-search engine Chomp."
}
```

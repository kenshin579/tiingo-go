# U.S. Fundamental Data API Documentation

> 출처: https://www.tiingo.com/documentation/fundamentals · 생성: 2026-09-04 (tools/gendocs)

## Fundamental Data API Documentation

**REST Endpoints**

Tiingo Fundamental Data API Endpoint is as an [add-on subscription](https://api.tiingo.com/products/fundamental-data-api). The DOW 30 are available for free/evaluation. If you would like access to all tickers as an add-on subscription, please E-mail [support@tiingo.com](mailto:support@tiingo.com)

```
# To see the available fundamental metrics, use this endpoint
https://api.tiingo.com/tiingo/fundamentals/definitions

# To request historical statement data, use this endpoint (ticker for active symbols)
https://api.tiingo.com/tiingo/fundamentals/<ticker>/statements

# For delisted or recycled symbols, use Tiingo permaTicker (stable identifier)
https://api.tiingo.com/tiingo/fundamentals/<permaTicker>/statements

# To request historical daily fundamental data, use this endpoint (ticker for active symbols)
https://api.tiingo.com/tiingo/fundamentals/<ticker>/daily

# For delisted or recycled symbols, use Tiingo permaTicker (stable identifier)
https://api.tiingo.com/tiingo/fundamentals/<permaTicker>/daily

# To request fundamental meta data, use this endpoint
https://api.tiingo.com/tiingo/fundamentals/meta
```

## 2.8.1 Overview

Tiingo offers a fundamental data API via a third-party provider. Data goes back over 20 years and covers US Equities and ADRs. We have used this provider extensively over the years and found them to be reputable, reliable, and accurate.

Benefits of Tiingo Fundamental Data API

- 5,500+ Equities Covered.
- Data includes Cash Flow, Income Statement, Balance Sheet, and Metric Data.
- Data is covered daily, quarterly, and for annual statements.
- Data is structured and normalized across a pre-define set of tickers for easy backtests and analysis.
- Data is served via a REST API API.
- Data is updated usually within 12-24 hours of being made available by the SEC.
- 3 Years of the DOW 30 tickers are available for free/evaluation.

## 2.8.2 Definitions Data

This endpoint can be used to check which the various fields available in the fundamentals endpoint. As we add more indicators, the output of this endpoint will change to reflect the addition of indicators.

**To request fundamental definitions REST endpoints.**

```
# To see the available fundamental metrics, use this endpoint
https://api.tiingo.com/tiingo/fundamentals/definitions
```

### Request

| FIELD NAME | PARAMETER | JSON FIELD | DATA TYPE | REQUIRED | DESCRIPTION |
| --- | --- | --- | --- | --- | --- |
| Tickers | GET | tickers | string | N | Specific tickers to return meta data for. If no string passed, will return meta data for all available tickers. Can either be a single ticker, a comma-separated list of tickers, or an array of strings (string[]). |
| Response Format | GET | format | string | N | Sets the response format of the returned data. Acceptable values are "csv" and "json". Defaults to JSON. |

### Response

| FIELD NAME | JSON FIELD | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Data Code | dataCode | string | An identifier representing the fundamentals field the value belongs to. For example, "peRatio." |
| Name | name | string | A human-friendly readable name of the field. |
| Description | description | string | A description of the field. |
| Statement Type | statementType | string | One of four values ("balanceSheet", "incomeStatement", "cashFlow", "overview"), representing which statement this value belongs to. For example, "netIncome" would belong to "incomeStatement". |
| Units | units | string | The unit the field value is in. Value is either "$", "%" or blank. If blank, value may either be an integer (like shares outstanding), or a ratio. |

### Example

```python
import requests
headers = {
    'Content-Type': 'application/json'
}
requestResponse = requests.get("https://api.tiingo.com/tiingo/fundamentals/definitions?token=<TOKEN>", headers=headers)
print(requestResponse.json())
```

**Response:**

```json
[
    {
        "dataCode":"liabilitiesCurrent",
        "name":"Current Liabilities",
        "description":"Debt or liabilities that are due within a year",
        "statementType":"balanceSheet",
        "units":"$"
    },
    {
        "dataCode":"rps",
        "name":"Revenue Per Share",
        "description":"Revenue per share",
        "statementType":"overview",
        "units":"$"
    }
    ...
]
```

## 2.8.3 Statement Data Endpoint

This endpoint returns data that is extracted from quarterly and annual statements.

For Tiingo statement data, use the endpoints below. Provide either a ticker (active symbols) or a Tiingo permaTicker (stable identifier for delisted/recycled symbols).

With this endpoint, the JSON (default) and CSV formats do share some differences in data strcture. The JSON data is nested, whereas the CSV format, by necessity, is a 2-D, flat structure. The data is the same, so choose a format that is simplest for you to digest.

Additionally, please ensure your code can handle new indicators as we will be continually updating these end points as we obtain new information and add new metrics.

**To request historical statement data for a stock, use the following REST endpoint.**

```
# To request historical statement data, use this endpoint (ticker for active symbols)
https://api.tiingo.com/tiingo/fundamentals/<ticker>/statements

# For delisted or recycled symbols, use Tiingo permaTicker (stable identifier)
https://api.tiingo.com/tiingo/fundamentals/<permaTicker>/statements

# To request historical statement data limited by date range, use this endpoint
https://api.tiingo.com/tiingo/fundamentals/<ticker>/statements?startDate=2019-06-30
```

### Request

| FIELD NAME | PARAMETER | JSON FIELD | DATA TYPE | REQUIRED | DESCRIPTION |
| --- | --- | --- | --- | --- | --- |
| Ticker (or Tiingo PermaTicker) | URL | N/A | string | Y | Required. Use the ticker for active symbols, or Tiingo PermaTicker (stable identifier) for delisted/recycled symbols. |
| As-Reported | GET | asReported | boolean | N | When false (default), the endpoint will return the most recent data, including any revisions for the reporting period. The dates will correspond to the fiscal end of the quarter or year. When true, the endpoint will return the data as it was reported on the release date. Similarly, the date will correspond to the date the filings were posted on the SEC. |
| Start Date | GET | startDate | date | N | This filter limits metrics to on or after the startDate (>=). Parameter must be in YYYY-MM-DD format. |
| End Date | GET | endDate | date | N | This filter limits metrics to on or before the endDate (<=). Parameter must be in YYYY-MM-DD format. |
| Sort | GET | sort | string | N | This field allows you to specify the sort direct and which column to sort by. Prepend "-" if you want descending order, otherwise it will be ascending. E.g. sort=date will sort by date in ascending order. sort=-date will sort by date in descending order. NOTE: For Fundamentals statement only the "date" field may be sorted upon |
| Response Format | GET | format | string | N | Sets the response format of the returned data. Acceptable values are "csv" and "json". Defaults to JSON. |

### Response

| FIELD NAME | JSON FIELD | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Date | date | datetime | The date the statement data was released to the public. |
| Quarter | quarter | int32 | An integer corresponding to the fiscal quarter reported. A value of "0" means this is an Annual Report. A value of "1" through "4" corresponds to the respective fiscal quarter. |
| Year | year | int32 | An integer corresponding to the fiscal year reported. |
| Statement Data | statementData | object | Statement data that is broken out by four different fields in a JSON object. Each fields corresponds to a statement. The four fields are:<br>balanceSheet: Balance sheet data<br>incomeStatement: Income Statement Data<br>cashFlow: Cash flow statement data<br>overview: Metrics and ratios that may be a combination of the fields from various statements<br>Each field will return an array of "dataCode" and "value" pairs that are described in the "Statement Data Table" below. |

To see what format of the array of fields are returned in the "statementData" field, please see the table below.

#### Statement Data Field

| FIELD NAME | JSON FIELD | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Data Code | dataCode | string | An identifier representing the fundamentals field the value belongs to. For example, "peRatio." |
| Value | value | float | The value of the field corresponding to the dataCode |

### Example

```python
import requests

headers = {
    'Content-Type': 'application/json'
}
requestResponse = requests.get("https://api.tiingo.com/tiingo/fundamentals/msft/statements?token=<TOKEN>", headers=headers)
print(requestResponse.json())
```

**Response:**

```json
[
    {
        "date":"2019-12-31",
        "quarter":4,
        "year":2019,
        "statementData":
            "balanceSheet":[
                {
                    "dataCode":"assetsCurrent",
                    "value":167074000000.0
                },
                {
                    "dataCode":"totalAssets",
                    "value":282794000000.0
                },
                {
                    "dataCode":"acctPay",
                    "value":8811000000.0
                },
                {
                    "dataCode":"assetsNonCurrent",
                    "value":115720000000.0
                },
                {
                    "dataCode":"accoci",
                    "value":-255000000.0
                },
                {
                    "dataCode":"totalLiabilities",
                    "value":172685000000.0
                },
                {
                    "dataCode":"taxLiabilities",
                    "value":31663000000.0
                },
                {
                    "dataCode":"taxAssets",
                    "value":0.0
                },
                {
                    "dataCode":"deferredRev",
                    "value":31221000000.0
                }
            ],
            "cashFlow": [...],
            "incomeStatement": [...],
            "overview": [...]
        }
    }
]
```

## 2.8.4 Daily Data

While statement data covers quarterly and annual reporting, some metrics that rely on price update daily, for example Market Capitalization, P/E Ratios, P/B Ratios, etc. This endpoint covers daily metrics.

For Tiingo daily fundamentals, use the endpoints below. Provide either a ticker (active symbols) or a Tiingo permaTicker (stable identifier for delisted/recycled symbols).

This endpoint is different than others in that the different daily metrics are "columns". Please note that we will continue to add new daily metrics, so the fields will change throughout time. We recommend you **do not** make parsing code that requires columns or fields to be in a particular order. If you do require this, please use the **columns** request parameter to ensure constant output, even if we add columns. For example, **columns=marketCap,peRatio** will always ensure only those two fields are returned in that exact order.

**To request daily metric data REST endpoints.**

```
# To request daily metric data, use this endpoint (ticker for active symbols)
https://api.tiingo.com/tiingo/fundamentals/<ticker>/daily

# For delisted or recycled symbols, use Tiingo permaTicker (stable identifier)
https://api.tiingo.com/tiingo/fundamentals/<permaTicker>/daily
```

### Request

| FIELD NAME | PARAMETER | JSON FIELD | DATA TYPE | REQUIRED | DESCRIPTION |
| --- | --- | --- | --- | --- | --- |
| Ticker (or Tiingo PermaTicker) | URL | N/A | string | Y | Required. Use the ticker for active symbols, or Tiingo PermaTicker (stable identifier) for delisted/recycled symbols. |
| Start Date | GET | startDate | date | N | This filter limits metrics to on or after the startDate (>=). Parameter must be in YYYY-MM-DD format. |
| End Date | GET | endDate | date | N | This filter limits metrics to on or before the endDate (<=). Parameter must be in YYYY-MM-DD format. |
| Sort | GET | sort | string | N | This field allows you to specify the sort direct and which column to sort by. Prepend "-" if you want descending order, otherwise it will be ascending. E.g. sort=date will sort by date in ascending order. sort=-date will sort by date in descending order. |
| Response Format | GET | format | string | N | Sets the response format of the returned data. Acceptable values are "csv" and "json". Defaults to JSON. |

### Response

| FIELD NAME | JSON FIELD | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Date | date | datetime | The date the daily data corresponds to. |
| Market Cap | marketCap | float | The value of the field corresponding to the market capitalization. |
| Enterprise Value | enterpriseVal | float | The value of the field corresponding to the enterprise value. |
| P/E Ratio | peRatio | float | The value of the field corresponding to the price/earnings ratio. |
| P/B Ratio | pbRatio | float | The value of the field corresponding to the price/book ratio. |
| Trailing PEG Ratio (1Y) | trailingPEG1Y | float | The value of the field corresponding to the trailing 1 year PEG Ratio. |

### Example

```python
import requests
headers = {
    'Content-Type': 'application/json'
}
requestResponse = requests.get("https://api.tiingo.com/tiingo/fundamentals/aapl/daily?token=<TOKEN>", headers=headers)
print(requestResponse.json())
```

**Response:**

```json
[
...
    {
        "date":"2020-05-08T00:00:00.000Z",
        "marketCap":1344207313550.0,
        "enterpriseVal":1413540313550.0,
        "peRatio":24.0971250971,
        "pbRatio":17.302742906,
        "trailingPEG1Y":5.4108999082
    },
    {
        "date":"2020-05-11T00:00:00.000Z",
        "marketCap":1365358868350.0,
        "enterpriseVal":1434691868350.0,
        "peRatio":24.4763014763,
        "pbRatio":17.5750073931,
        "trailingPEG1Y":5.4960422406
    }
]
```

## 2.8.5 Meta Data

This endpoint can be used to check which tickers have been updated with new fundamental data.

Like the endpoints above, as we add new meta data about companies and their fundamentals, this endpoint will change output. Please do not code with the assumption that column orders will always be maintained. If you must, please pass the **columns** parameter to ensure the output maintains its column order. For example, **columns=ticker,name** will always ensure those columns are returned in that exact order.

**To request meta data REST endpoints.**

```
# To request fundamental meta data, use this endpoint
https://api.tiingo.com/tiingo/fundamentals/meta
```

### Request

| FIELD NAME | PARAMETER | JSON FIELD | DATA TYPE | REQUIRED | DESCRIPTION |
| --- | --- | --- | --- | --- | --- |
| Tickers | GET | tickers | string | N | Specific tickers to return meta data for. If no string passed, will return meta data for all available tickers. Can either be a single ticker, a comma-separated list of tickers, or an array of strings (string[]). |
| Response Format | GET | format | string | N | Sets the response format of the returned data. Acceptable values are "csv" and "json". Defaults to JSON. |

### Response

| FIELD NAME | JSON FIELD | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Perma Ticker | permaTicker | string | Permanent Tiingo Ticker mapping to the security. Can be used as a primary key. |
| Ticker | ticker | string | Ticker related to the asset. |
| Name | name | string | Full-length name of the asset. |
| Is Active | isActive | boolean | Boolean describing whether or not the ticker is still actively traded. If false, this ticker is delisted. |
| Is ADR | isADR | boolean | Boolean describing whether or not the ticker is an ADR. Value is true if listed ticker is an ADR. |
| Sector | sector | string | Sector information that is derived from sicSector. |
| Industry | industry | string | Sector information that is derived from sicIndustry. |
| SIC Code | sicCode | int32 | SIC Code that represents company's business activities. |
| SIC Sector | sicSector | string | Sector as determined by the SIC Code. |
| SIC Industry | sicIndustry | string | Industry as determined by the SIC Code. |
| Reporting Currency | reportingCurrency | string | The currency the company reports their SEC statement filings in. |
| Location | location | string | Location/domicile of the company. States are included for U.S. companies, otherwise countries for non-US companies. |
| Company Website | companyWebsite | string | The website of the company when available. |
| SEC Filing Website | secFilingWebsite | string | A URL to where you can find the company's SEC filings directly on the SEC website. |
| Statement Last Updated | statementLastUpdated | datetime | The timestamp the statement data/endpoint was last updated for the ticker. |
| Daily Last Updated | dailyLastUpdated | datetime | The timestamp the daily data/endpoint was last updated for the ticker. |

### Example

```python
import requests
headers = {
    'Content-Type': 'application/json'
}
requestResponse = requests.get("https://api.tiingo.com/tiingo/fundamentals/meta?token=<TOKEN>", headers=headers)
print(requestResponse.json())
```

**Response:**

```json
[
    {
        "permaTicker":"US000000000253",
        "ticker":"bac",
        "name":"Bank of America Corp",
        "isActive":true,
        "sector":null,
        "statementLastUpdated":"2019-12-22T22:08:11.534Z",
        "dailyLastUpdated":"2019-12-22T22:08:17.530Z"
    },
    {
        "permaTicker":"US000000000133",
        "ticker":"ibm",
        "name":"International Business Machines Corp",
        "isActive":true,
        "sector":null,
        "statementLastUpdated":"2019-12-05T20:01:58.748Z",
        "dailyLastUpdated":"2019-12-22T22:08:17.530Z"
    }
...
]
```

## 2.8.6 Additional Information & FAQ

We decided to partner with this data provider as over the years we have found their data process to be incredibly rigorous, their mission aligned with ours, and their support incredibly responsive. We hope the follow FAQ will be helpful to you, if you have any more questions please E-mail us at [Support@tiingo.com](mailto:support@tiingo.com?subject=Fundamentals%20Data%20Question)

#### Is the Statements data processed by hand or programmatically?

Data collection goes through a mixture of machine curation and human oversight.

#### What SEC filings are used for Fundamental Data?

Also sourcing includes the following docs: 10-12, S-11, S-4, 10SB and A-1 (Domestic), 20-F and 20FR (ADR), 40-F and 40FR (Canadian), F-1, F-10 and F-4 (ADR and Canadian), and limited 6-K (for ADR and Canadian). 8-K is being considered as well, and more others may be added in the future.

#### Do you use XBRL?

We have found XBRL unreliable and so do not rely on it. Human readable documents are a better measure of what companies intend to convey and so we focus on interpreting these.

#### Is prior period data pulled from the latest report? Or is only the current period taken? As in 2018 annual info is from 2019 annual filings?

Prior period data is pulled from the latest report for the Most-Recent (asReported=false) dimension, current period is taken only for As-Reported (asReported=true) dimension.

#### What accuracy checks are performed?

Our big focus is on reconciling to the underlying human readable SEC document. In addition we deploy a variety of human and algorithmic checks.

#### How quickly is data updated after a company reports earnings?

Less than 24 hours from the SEC form 10 filing.

#### What currency is data returned in?

All values are returned in USD. If a company reports in a different currency, the values are converted to USD using a time-appropriate FX-rate before being served to you via the API. This is especially relevant to ADRs who often report in a different currency than USD.

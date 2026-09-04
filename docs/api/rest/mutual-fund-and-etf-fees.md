# Current and Historical Mutual Fund and ETF Fees

> 출처: https://www.tiingo.com/documentation/mutual-fund-and-etf-fees · 생성: 2026-09-04 (tools/gendocs)

## Mutual Fund and ETF Fee Data API Documentation

**REST Endpoints**

This endpoint is for enterprise and institutional clients only. Please reach out to [Sales@tiingo.com](mailto:sales@tiingo.com) for licensing and pricing.

```
# To obtain top-level fund data, including description and share classes, use the following REST endpoint
https://api.tiingo.com/tiingo/funds/<ticker>

# To obtain detailed current and historical fee data, use the following REST endpoint
https://api.tiingo.com/tiingo/funds/<ticker>/metrics
```

## 2.9.1 Overview

Tiingo tracks and processes official Mutual Fund & ETF Fee data from over 36,000 Mutual Funds and ETFs. The data includes a detailed breakdown data of fees, even including custom fees (like check processing fee if withdrawing money via check).

Benefits of Mutual Fund & ETF Fee Data

- 36,000+ Mutual Funds and ETFs covered.
- Data includes historical as well as current data.
- Fee data is detailed and broken down (e.g. management fee, 12b-1 fees, and more).
- Fee data is updated intraday as new data comes online from Fund Companies.
- Multiple share classes are mapped allowing you to easily compare fees across share classes.

## 2.9.2 Fund Overview

**To obtain top-level fund data, including description and share classes, use the following REST endpoint.**

```
# To obtain top-level fund data, including description and share classes, use the following REST endpoint
https://api.tiingo.com/tiingo/funds/<ticker>
```

### Request

| FIELD NAME | PARAMETER | JSON FIELD | DATA TYPE | REQUIRED | DESCRIPTION |
| --- | --- | --- | --- | --- | --- |
| Ticker | URL or GET | tickers | string | Y | Ticker related to the fund. |

### Response

| FIELD NAME | JSON FIELD | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Ticker | ticker | string | Ticker related to the fund. |
| Name | name | string | Full-length name of the fund. |
| Description | description | string | Long-form descripton of the fund. |
| Share Class | shareClass | string | Share class of the fund as described by the parent fund company. |
| Net Expense Ratio | netExpense | float | The top-level net expense ratio for the fund. |
| Other Share Classes | otherShareClasses | object[] | An array of objects representing related share classes of the given fund. See below for the object defintion table. |

To see what fields are returned in the "otherShareClasses" field, please see the table below.

| FIELD NAME | JSON FIELD | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Ticker | ticker | string | Ticker related to the fund. |
| Name | name | string | Full-length name of the fund. |
| Share Class | shareClass | string | Share class of the fund as described by the parent fund company. |
| Net Expense Ratio | netExpense | float | The top-level net expense ratio for the fund. |

### Example

```python
import requests
headers = {
    'Content-Type': 'application/json'
}
requestResponse = requests.get("https://api.tiingo.com/tiingo/funds/vfinx?token=<TOKEN>", headers=headers)
print(requestResponse.json())
```

**Response:**

```json
{
    "ticker":"vfinx"
    "name":"VANGUARD 500 INDEX FUND INVESTOR SHARES",
    "shareClass":"INVESTOR SHARES",
    "description":"The Fund employs an indexing investment approach designed to track the performance of the Standard & Poor's 500 Index, a widely recognized benchmark of U.S. stock market performance that is dominated by the stocks of large U.S. companies. The Fund attempts to replicate the target index by investing all, or substantially all, of its assets in the stocks that make up the Index, holding each stock in approximately the same proportion as its weighting in the Index.",
    "netExpense":0.0014,
    "otherShareClasses":[
    {
        "ticker":"VFIAX",
        "name":"VANGUARD 500 INDEX FUND ADMIRAL SHARES",
        "shareClass":"ADMIRAL SHARES",
        "netExpense":0.0004
    },
    {
        "ticker":"VFFSX",
        "name":"VANGUARD 500 INDEX FUND INSTITUTIONAL SELECT SHARES",
        "shareClass":"INSTITUTIONAL SELECT SHARES",
        "netExpense":0.0001
    },
    {
        "ticker":"VOO",
        "name":"Vanguard S&P 500 ETF",
        "shareClass":"ETF SHARES",
        "netExpense":0.0003
    }
]
```

## 2.9.3 Historical and Current Mutual Fund & ETF Fee Data

**To obtain detailed current and historical fee data, use the following REST endpoint.**

```
# To obtain detailed current and historical fee data, use the following REST endpoint
https://api.tiingo.com/tiingo/funds/<ticker>/metrics
```

### Request

| FIELD NAME | PARAMETER | JSON FIELD | DATA TYPE | REQUIRED | DESCRIPTION |
| --- | --- | --- | --- | --- | --- |
| Ticker | URL | N/A | string | Y | Ticker related to the asset. |
| Start Date | GET | startDate | date | N | If startDate or endDate is not null, historical data will be queried. This filter limits metrics to on or after the startDate (>=). Parameter must be in YYYY-MM-DD format. |
| End Date | GET | endDate | date | N | If startDate or endDate is not null, historical data will be queried. This filter limits metrics to on or before the endDate (<=). Parameter must be in YYYY-MM-DD format. |
| Resample Freq | GET | resampleFreq | string | N | his allows you to set the frequency in which you want data resampled. For example "1hour" would return the data where OHLC is calculated on an hourly schedule. The minimum value is "1min". Both units in minutes (min) and hours (hour) are accepted. Format is # + (min/hour); e.g. "15min" or "4hour". If no value is provided, defaults to 5min. |
| Response Format | GET | format | string | N | Sets the response format of the returned data. Acceptable values are "csv" and "json". Defaults to JSON. |

### Response

| FIELD NAME | JSON FIELD | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Prospectus Date | prospectusDate | date | The prospectus date when the corresponding fund expense data was published. |
| Net Expense Ratio | netExpense | float | Fund's net expense ratio, or the net expenses related to the fund. |
| Gross Expense Ratio | grossExpense | float | Fund's gross expense ratio, or the expenses related to running the fund. |
| Management Fee | managementFee | float | Fund's management fee, or the fees paid to the manager and/or advisors. |
| 12b-1 Fee | 12b1 | float | Fund's fee related to marketing expenses. |
| Non-12b-1 Fee | non12b1 | float | Fund's fee related to distribution and smilar non 12b-1 fees. |
| Other Fund Expenses | otherExpenses | float | Fund's other expenses, or expenses related to legal, adminstrative, custodial, etc. |
| Acquired Fund Fees | acquiredFundFees | float | Fund's acquired fund fees, or expenses related to underlying businesses or funds. |
| Fee Waiver | feeWaiver | float | Fund's fee waiver, or discount on fees. |
| Exchange Fee (USD) | exchangeFeeUSD | float | Fund's exchange fee if charged in USD, or expenses related to exchanging or transferring funds to another fund in the fund's family. |
| Exchange Fee (%) | exchangeFeePercent | float | Fund's exchange fee if charged as a percentage, or expenses related to exchanging or transferring funds to another fund in the fund's family. |
| Front Load Fee | frontLoad | float | Fund's front load fee, or the upfront fee charged when investing in the fund. |
| Back Load Fee | backLoad | float | Fund's back load fee, or the back-end fee charged when redeeming from the fund. |
| Dividend Load Fee | dividendLoad | float | Dividend load fee, or charges on reinvested dividends. |
| Shareholder Fee | shareholderFee | float | Fund's shareholder fees, or the potential fees when buying/selling a fund. |
| Account Fee (USD) | accountFeeUSD | float | Fund's account fees if charged in USD, or the fee required to maintain your account in USD. |
| Account Fee (%) | accountFeePercent | float | Fund's account fees if charged as a percentage, or the fee required to maintain your account in percentage terms. |
| Redemption Fee (USD) | redemptionFeeUSD | float | Fund's redemption fees if charged in USD, or the fee charged if funds are redeemed early (as defined by the fund company). |
| Redemption Fee (%) | redemptionFeePercent | float | Fund's redemption fees as a percentage, or the fee charged if funds are redeemed early (as defined by the fund company). |
| Portfolio Turnover | portfolioTurnover | float | Portfolio turnover. |
| Miscellaneous Fees | miscFees | float | Fund's miscellaneous fees. |
| Custom Fees | customFees | object[] | Fund's custom fees. For a full breakdown, see the table below. |

To see what fields are returned in the "customFees" field, please see the table below.

| FIELD NAME | JSON FIELD | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Label | label | string | Label related to the custom fee. |
| Value | value | float | Value of the custom fee field. |
| Units | units | char | "$" if the value is in dollars or "%" if the value is in percentage terms. |
| Parent Fee | parentFee | string | The parent fee the custom fee's belongs under. |

### Example

```python
import requests

headers = {
    'Content-Type': 'application/json'
}
requestResponse = requests.get("https://api.tiingo.com/tiingo/funds/berix/metrics?token=<TOKEN>", headers=headers)
print(requestResponse.json())
```

**Response:**

```json
[
    {
        "prospectusDate":"2021-03-01",
        "netExpense":0.0049,
        "grossExpense":0.0061,
        "managementFee":0.004,
        "12b1":0.0,
        "non12b1":0.0,
        "otherExpenses":0.0021,
        "acquiredFundFees":0.0,
        "feeWaiver":-0.0012,
        "exchangeFeeUSD":0.0,
        "exchangeFeePercent":0.0,
        "backLoad":0.0,
        "frontLoad":0.0,
        "dividendLoad":0.0,
        "shareholderFee":15.0,
        "accountFeeUSD":0.0,
        "accountFeePercent":0.0,
        "redemptionFeeUSD":0.0,
        "redemptionFeePercent":0.0,
        "portfolioTurnover":0.63,
        "miscFees":45.0,
        "customFees":[
            {
                "label":"Wire Fee",
                "value":20.0,
                "units":"$",
                "parentFee":"miscFees"
            },
            {
                "label":"Overnight check delivery fee",
                "value":25.0,
                "units":"$",
                "parentFee":"miscFees"
            }
        ],

    }
    ...
]
```

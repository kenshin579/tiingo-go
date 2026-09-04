# Search API Documentation

> 출처: https://www.tiingo.com/documentation/utilities/search · 생성: 2026-09-04 (tools/gendocs)

## Endpoints

**REST Endpoints**

Search Endpoint has just been launched and is in early beta. Responses objects are subject to change. We do not recommend building production code using this endpoint while in beta.

```
# Search Endpoint
https://api.tiingo.com/tiingo/utilities/search/<query>
# or
https://api.tiingo.com/tiingo/utilities/search?query=<query>
```

## 4.1.1 Overview

Tiingo's search feature lets you find specific assets in our database by the ticker or the name of the asset. This endpoint lets you segment by active, delisted, tickers across asset classes. The endpoint first searches for ticker matches and then expands to matches in the name of the asset.

This endpoint is useful for looking up existing assets

## 4.1.2 Search Endpoint

**To request search query data, use the following REST endpoints.**

```
# Search Tiingo database for specific assets
https://api.tiingo.com/tiingo/utilities/search/apple
# or
https://api.tiingo.com/tiingo/utilities/search?query=apple
```

### Request

| FIELD NAME | PARAMETER | JSON FIELD | DATA TYPE | REQUIRED | DESCRIPTION |
| --- | --- | --- | --- | --- | --- |
| Search Query | GET | query | string | Y | The search query to use to look up assets. |
| Exact Ticker Match | GET | exactTickerMatch | boolean | N | True to only include exact ticker matches based on the search query. If set to true, no partial matches will be included and asset names will not be searched. |
| Include Delisted | GET | includeDelisted | boolean | N | True to include delisted tickers and false (default) to exclude delisted tickers. |
| Limit | GET | limit | int32 | N | The maximum number of assets to return. Defaults to 10 and can be set to a maximum of 100. |
| Response Format | GET | format | string | N | Sets the response format of the returned data. Acceptable values are "csv" and "json". Defaults to JSON. |
| Columns | GET | columns | string[] | N | Allows you to specify which columns you would like returned from the output. Pass an array of strings of column names to get only this columns back. |

### Response

| FIELD NAME | JSON FIELD | DATA TYPE | DESCRIPTION |
| --- | --- | --- | --- |
| Ticker | ticker | string | Ticker of the given asset. |
| Name | name | string | The name of the asset. |
| Asset Type | assetType | string | The asset type of the asset (Stock, ETF, & Mutual Fund). |
| Is Active | isActive | boolean | True if the ticker is still actively quoted, and false if the ticker is no longer actively quoted (delisted). |
| Tiingo PermaTicker | permaTicker | string | Placeholder for an upcoming change to the Tiingo API that allows querying by permaticker. |
| OpenFIGI Ticker | openFIGI | string | Placeholder for an upcoming change to the Tiingo API that allows querying by the openFIGI ticker. |

### Example

```python
import requests
headers = {
    'Content-Type': 'application/json'
}
requestResponse = requests.get("https://api.tiingo.com/tiingo/utilities/search?query=apple&token=<TOKEN>", headers=headers)
print(requestResponse.json())
```

**Response:**

```json
[
    {
    "ticker":"AAPL",
    "assetType":"Stock",
    "countryCode":"US",
    "isActive":true,
    "name":"Apple Inc",
    "openFIGI":"BBG000B9XRY4",
    "permaTicker":"US000000000038"
    },
    {
    "ticker":"PNPL",
    "assetType":"Stock",
    "countryCode":"US",
    "isActive":true,
    "name":"Pineapple Exprss",
    "openFIGI":null,
    "permaTicker":"US000000047877"
    },
..]
```

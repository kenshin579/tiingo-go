# Symbology API Documentation

> 출처: https://www.tiingo.com/documentation/appendix/symbology · 생성: 2026-09-04 (tools/gendocs)

## 5.3 Symbology

This guide is covers Tiingo's symbol format for querying different assets throughout the API.

## 5.3.1 Introduction

Tiingo follows a symbology format to help disambiguate share classes, preferreds, and other types of assets. This section will be expanded as we ready to expand our API into permatickers and delisted ticker support.

## 5.3.2 Equities

Right now the Tiingo API supports Equity, Mutual Fund, and ETF prices for US markets and Equity Prices for Chinese markets.

Since periods are often used in URLs to signify file extensions, we do not use periods in our symbol names given our REST APIs. Instead of periods ("."), we use hypens ("-"). This means "BRK.A" is "BRK-A" within Tiingo.

All tickers that have separate share classes follow the format:

```json
{SYMBOL}-{SHARE CLASS}
```

For example Berkshire Hathaway Class A shares are "BRK-A".

Similarly, for preferred shares, you may use the format:

```json
{SYMBOL}-P-{SHARE CLASS}
```

For example Simon Property Group's Preferred J series shares would be "SPG-P-J".

Mutual funds and Closed-End-Funds (CEFs) follow their typical format. Mutual Funds usually end with the letter "X" (e.g. "VFINX") and CEFs begin and end with the letter "X" (e.g. "XAIFX").

Tiingo supports delisted data for tickers that have not yet been recycled. You may see some delisted data that ends in a digit. These are typically tickers that are stored in our database, but we do not yet have price data. As Tiingo expands, we will expand the offering into even older delisted data.

## 5.3.3 CryptoCurrency & FX

Tiingo's methodology for CryptoCurrencies and FX currencies follow the same symbol method.

Because forward slashes ("/") are used to separate locations within URLs and directories, Tiingo does not use forward slashes within currency symbols to prevent URL amibiguity. Instead we remove the forward slashes. Therefore the format for currencies is:

```json
{BASE_CURRENCY}{QUOTE_CURRENCY}
```

For "EUR/USD" the base currency is "EUR" and the quote currency is "USD". This means the Tiingo symbol is "EURUSD".

For "BTC/ETH" the base currency is "BTC" and the quote currency is "ETH". This means the Tiingo symbol is "BTCETH".

For crypto DEX state prices, you will see the following format:

```
'tngosp' + {BASE_CURRENCY}{QUOTE_CURRENCY}
```

For example the state price for "mkreth" will be distributed as "tngospmkreth" (tngosp + mkreth) in our endpoints.

State prices are the implied current price of a pair based on the DEX pool composition at the given moment. An imperfect, but somewhat helpful analogy would be the "mid" price of a DEX pool.

For Crypto Synthetics, you will see the following format:

```json
{BASE_CURRENCY}/{QUOTE_CURRENCY}
```

For example the crypto synthetic price for BTC-USD would be BTC/USD.

Synthetics are different than normal prices, and unless you are on the institutional plans, you will not see them on any public endpoints. They are Benchmark rates that take all eligible exchange crosspairs of the BASE_CURRENCY and then FX route all pairs into the QUOTE_CURRENCY and weight them. E.g. btcusdc, btcusd, btcusdt, ...btceur would all be converted into usd using intermediary fx pairs (with the exception of btcusd) and then weighted to produce a benchmark price. This way if the fx pair, or quote currency, change rapidly - the prices can update even before the base currency trades.

Synthetics are offered in three flavors, and all follow the above symbol format:

1. VWAP Synthetics: VWAP Benchmark rates (CEX and DEX)
2. Top-of-Book Synthetics: Top-of-Book Benchmark rates (CEX and Perp DEX)
3. State Synthetics: State Benchmark rates (DEX and Perp DEX)

# Tiingo API Changelog Documentation

> 출처: https://www.tiingo.com/documentation/general/changelog · 생성: 2026-09-04 (tools/gendocs)

## 1.3 Changelog

Use this page to keep up with the latest additions and changes to the API.

## 2026-07-21

### BOATS Overnight Real-time REST and BOATS Overnight Real-time WebSocket - New Product (Beta)

- We've added a new BOATS Overnight Real-time API (currently in beta) that delivers venue-native Blue Ocean ATS (BOATS) equity market data during the overnight session (8:00 PM–3:59 AM ET).
- Includes Current Top-of-Book & Last Price REST endpoints and Historical Overnight Intraday Prices under `/boats`, plus a real-time WebSocket firehose at `wss://api.tiingo.com/boats`. See the [REST documentation](https://www.tiingo.com/documentation/boats) and [WebSocket documentation](https://www.tiingo.com/documentation/websockets/boats).
- Coverage spans 12,000+ quotable US equities with top-of-book quotes (bid/ask/mid), last trade price and size, and overnight intraday OHLC bars.
- Access requires the BOATS Real-time add-on entitlement (+$9/month on Power and Commercial plans; unavailable on Free tier). More on the [BOATS product page](https://www.tiingo.com/products/boats-blue-ocean-ats-real-time-overnight-stock-prices).

## 2026-07-07

### Equity Realtime REST and Equity Realtime WebSocket - New Product (Beta)

- We've added a new Equity Realtime API (currently in beta) that builds consolidated realtime equity snapshots from multiple Equity venues (exchanges, ATS, and OTC venues). Tiingo is the first company to make these exchange-compliant derived metrics available.
- Includes a Liquidity Spread Top-of-Book & Reference Price endpoint and a Historical Intraday Prices endpoint, both served under `/tiingo/equity/intraday`. Market hours run from 4am ET to 8pm ET with liquidity scored bid, ask, last sale, Tiingo reference price, and day-level OHLC fields.
- Historical intraday OHLCV bars support configurable minute and hourly resampling, optional after-hours data, and force-filled bars for charting workflows.
- Data is served via both [REST endpoints](https://www.tiingo.com/documentation/equity-realtime-stock-data) and [WebSockets](https://www.tiingo.com/documentation/websockets/equity-realtime-stock-data).

### EOD Composite Feeds

- Better tracking of expanded upcoming corporate actions, including a massive internal refactor of how we process these and also distribute upcoming values.
- Added two new cleansing checks - including a new type of check for more accurate error cleansing of subtle erroneous points.

### Crypto Synthetics - State Price

- Added new features to refine specific state price pool inclusion.

### Symbology

- Documented the symbology format for Crypto DEX State Prices, which use the `tngosp` prefix followed by the base and quote currencies (e.g. `tngospmkreth` for the mkreth state price).
- Documented the symbology format for Crypto Synthetics, which use a `{BASE}/{QUOTE}` convention (e.g. `BTC/USD`).
- Clarified the three Synthetic flavors: VWAP Synthetics (CEX and DEX), Top-of-Book Synthetics (CEX and Perp DEX), and State Synthetics (DEX and Perp DEX).

## 2026-05-12

### General

- Huge data center migration, massive boost to infra as we accelerate. More to come.
- Smoother backend improvements on subscription management.
- API Token rotation now available via the [Token Page](https://www.tiingo.com/account/api/token).

### Dividend Corporate Actions and EOD Composite Feeds

- Further large improvement in corp actions coverage across equities, ETFs, and mutual funds.

### IEX Feeds REST

- Improved reference price formula to prevent occasional outliers - resulting in smoother reference prices.
- Huge upgrade in IEX market performance.

### Crypto API & Benchmarks Data

- Improved resiliency of algorithmic calculations of Benchmarks.
- Improved responsiveness of RPC performance for faster DEX recovery in case of RPC recovery.
- Increased coverage of DEXes, now covering 160+ DEXes and CEXes.

### Crypto Benchmark Data

- Better FX Engine outlier detection.

### News API Data

- Added new sources.

## 2026-03-20

### Dividend Corporate Actions and EOD Composite Feeds

- Massive improvement in corp actions across equities, ETFs, and mutual funds. Now catch even more upcoming distributions with more detail.

### Fundamental Data

- Improved asReported=true functionality for Fundamentals Daily endpoints.

### Crypto API Data

- Further improved speed of recovery of blockchain outages and RPC failtures for swap tracking.

### Crypto Benchmark Data

- Better FX Engine outlier detection.

### News API Data

- Now cover 70+ million articles!
- Added new sources.

## 2026-01-31

### Fundamental Data

- Documentation now added highlighting query support of permaTicker for /statements and /daily Fundamental Endpoints.
- Fixed edge case calculations in certain ADRs that report quarterly.

### FX Feeds REST and FX Feeds WebSocket

- Added additional assets.

### Crypto API Data

- Improved recovery of blockchain outages and RPC failtures for swap tracking.
- Better handling of exchange recovery when order book data is corrupted exchange-side.

### Crypto Benchmark Data

- Calibrated low-latency configurations for thinly-traded assets.

## 2025-12-31

### General

- Automatically normalize symbols on websocket subscribes, so case sensitivity will no longer matter on remaining websocket connections.

### FX Feeds REST

- Better handling of date limits when querying FX data

### IEX Feeds REST

- Improved logic of derived calculation, resulting in more relevant updates.

### Crypto API Data

- Massively added coverage of new CEXes, DEXes, and blockchains. The product page is now updated with a list of 140+ exchanges (CEXes and DEXes covered) with native blockchain/DEX support.
- Better handling of exchange recovery when order book data is corrupted exchange-side.

### Crypto Benchmark Data

- Expansion of secondary symbols, allowing more exchange-specific benchmarks across a wider avenue of DEXes and CEXes.
- Improved Benchmark State Price algorithm for price stability.
- Created hybrid state and top-of-book algorithms for Perp DEXes.

## 2025-06-30

### General

- Automatically normalize symbols on websocket subscribes, so case sensitivity will no longer matter on remaining websocket connections.

### IEX Feeds WebSocket

- Improved derived calculation methodology to result in faster updates in symbols that receive less updates.

### Crypto API Data

- Improved crypto exchange error mitigation (i.e. faster handling of when exchanges go down)
- Massively added coverage of new CEXes, DEXes, and blockchains.

### Crypto Benchmark Data

- Added exchange-specific benchmark rates.
- Added new kind of contract prices (tracking mint and redeems of stablecoins - specifically Ethena).
- Expanded Solana coverage of DEX states and program versions.

## 2025-03-10

### Fundamental Data

- Added as-reported calculations to daily calculations.

### Crypto API Data

- Added coverage of new CEXes, DEXes, and blockchains.

### Crypto Benchmark Data

- Massively expanded DEX coverage universe including new chains. Contact sales@tiingo.com for more details.

## 2025-01-10

### General

- Improvements in WebSocket performance.

### IEX Feeds REST

- Added changes to handle IEX market data policy changes coming on 2025-02-01.

### IEX Feeds WebSocket

- Added a new thresholdLevel of 6 to create derived data streams (tngoLast) as part of the changes to handle IEX market data policy changes coming on 2025-02-01.

### Dividend Corporate Actions - Beta Product

- Improved coverage and detection of events.

### Split Corporate Actions - Beta Product

- Improved coverage and detection of events.

### Crypto API Data

- Added coverage of new CEXes and DEXes.
- Minor bug fixes.

### Crypto Benchmark Data

- Added State Price endpoints to returning streaming prices for DEX-dominated state prices. Includes both EVM DEXes and Solana DEXes. This is a great "low-latency" DEX Product. Contact sales@tiingo.com for more details.
- Added feature to explore underlying sources creating Benchmark rates for Top-of-Book Benchmarks and State Price Benchmarks.

## 2024-07-10

### General

- Improvements in WebSocket performance.

### Dividend Corporate Actions - Beta Product

- Improved coverage and detection of events.

### Split Corporate Actions - Beta Product

- Improved coverage and detection of events.

### News API Data

- Added new sources.

### Crypto API Data

- Improved error-checking of top of book data from exchanges that may send incorrect data.
- Added coverage of new CEXes and DEXes.

### Crypto Benchmark Data

- Added feature to explore underlying sources creating Benchmark rates.

## 2024-03-28

### General

- General performance improvements across several REST endpoints.
- Improvements in WebSocket performance and simultaenous connections.

### Dividend Corporate Actions - Beta Product

- Added ability to query data per ticker.
- Fixed timezone information on dates to allow dates to be accurate to the timezone of the event.
- Fixed several bugs.

### Split Corporate Actions - Beta Product

- Added ability to query data per ticker.
- Fixed timezone information on dates to allow dates to be accurate to the timezone of the event.
- Fixed several bugs.

### News API Data

- Improved accuracy of tagging.
- Better processing of News websites that may have issues in their structure.

### Crypto API Data

- Improved resiliency of top of book data from exchanges that may send incorrect data.

## 2024-01-04

### General

- Cleaned structure of several documentation and product pages.
- General performance improvements across several endpoints.
- Better handling of concurrent WebSocket connections for api tokens.
- Made page improvements on the Tiingo App.
- New Changelog process: Updates will be aggregated and reported at least quarterly.

### Dividend Corporate Actions - Beta Product

- Made several data improvements to capture more distributions.
- Now captures announcements throughout the day.
- Added new date fields, e.g. recordDate, paymentDate, declarationDate, etc.
- Documentation clarifications added.

### Split Corporate Actions - Beta Product

- Made several split data improvements to capture more splits.
- Now captures announcements throughout the day.

### Security Master Data - Enabled Commercial Clients

- Improved handling of various exchange actions for improved accuracy.

### EOD Composite Feeds

- Improved update performance.

### News API Data

- Expanded coverage and accuracy of tagging.
- Better & more efficient crawling of different types of news sources.

### Fund Fees

- Increased coverage to more Fund companies. Now supports over 41,000+ funds.

### Crypto API Data

- Added more crypto pairs that are being tracked.
- Expansion into new DEXes and CEXes.
- Expanded coverage of existing DEXes on Solana.
- Reintroduced Standard top-of-Book websocket with a subset of quality exchanges. This is a separate product than the Benchmark top-of-book websocket.
- Top-of-book Benchmark WebSocket now supports more quote currency conversion.
- Improved resiliency and handling of RPC load balancing and rotation for DEX aggregations.
- Tweaked general outlier detection algorithm.
- Fixed and improved outlier detection on Crypto Synthetic Benchmark WebSockets.
- Improved efficiency and speed of low latency top-of-book endpoints.

## 2023-09-24

### General

- Changed the default tiingo.com to direct to our API pages. The old main Tiingo.com has been redirected to the Tiingo App (https://app.tiingo.com).
- Changed usage and pricing tiers for Power and Free customers.
- Product and Documentation pages have been refined.
- General performance improvements across several endpoints.

### Dividend Corporate Actions - New Product

- We've added a new [Stock, ETF, and Mutual Fund Dividend API](https://www.tiingo.com/products/corporate-actions/stock-etf-mutual-fund-dividend-api) that lets you capture more detailed distribution data, for both historical, current, and future distributions that are coming up.

### Split Corporate Actions - New Product

- We've added a new [Stock, ETF, and Mutual Fund Split API](https://www.tiingo.com/products/corporate-actions/stock-etf-mutual-fund-split-api) that lets you capture more detailed splits data, for both historical, current, and future splits that are coming up.

### Fundamental Data

- Refined calculation and edge cases for fundamental data metrics.

### Crypto Feeds

- Added significantly more crypto pairs that are being tracked.
- Expansion into new DEXes.
- Improved efficiency and speed of low latency top-of-book endpoints.

## 2023-06-14

### General

- Product pages have been refined.
- General performance improvements across several endpoints.

### EOD Composite Feeds

- Improved Security Master data for customers who have access to the beta endpoint.

### Tiingo Fundamentals

- Fundamentals is now out of beta.
- Improved calculations for daily metrics.

### Tiingo Forex

- Forex is now out of beta.

### Crypto Feeds

- Added DEX tick storage - each event and event meta data is now available as a bulk download.
- Expansion into new DEXes.
- Added low latency top-of-book endpoints.

## 2023-01-10

### General

- Happy New Year!
- New Power Plans with new pricing. Existing customers will be kept on their existing plans.
- General performance improvements across several endpoints.

### EOD Composite Feeds

- Improved Security Master data for customers who have access to the beta endpoint.
- Improved Mutual Fund NAV data.

### Tiingo News

- Expanded news coverage into more ASX equities.
- Added even more news sources.
- Improved tagging on existing news sources.

### Tiingo Fundamentals

- Fundamental Daily Incremental Bulk Download is now enabled. Reach out to support for more information.

### Crypto Feeds

- Expansion into new CEXes, DEXes, and Blockchains.
- Added new tickers as well as improving speeds to keep data updating even during high stress environments.

## 2022-11-10

### General

- Major core engine improvements in IEX, Crypto, and FX data feeds to keep prices updating quickly even during high market volatility.
- Improvements in the WebSocket API..
- Added connectivity to a new futures exchange, TBA..
- We have migrated infrastructure to data centers closer to exchanges - including a preseence in the same data centers as many exchanges.
- New Logos! Check them out :)
- General performance improvements across several endpoints.

### IEX Feeds

- Significant performance improvements to keep the data fast and current for both REST and WebSocket APIs.
- Significant performance improvement in the WebSocket API in particular - especially during high market volatility periods.

### Tiingo News

- Added coverage for Australian equities. Reach out to [sales@tiingo.com](mailto:sales@tiingo.com) for more information on how to get connected.
- Improved tagging speeds.
- Millions of articles added and continue to be added.

### Tiingo Fundamentals

- Data improvements for meta data fields.
- Faster ingestion of data for API consumption when new financial values are made available.

### Crypto Feeds

- Expansion into new CEXes, DEXes, and Blockchains.
- Added new tickers as well as improving speeds to keep data updating even during high stress environments.

### FX Feeds

- Engine improvements to keep data updating quickly even during high market volatility.s.

## 2022-8-04

### General

- We have started the migration of infrastructure to data centers closer to exchanges - including a preseence in the same data centers as many exchanges.
- General performance improvements across several endpoints.

### Tiingo Fundamentals

- Significant performance improvements for REST queries.

### Crypto Feeds

- Improved outlier detection for Crypto Synthetics and Enterprise Clients who opt-in to data pruning.
- Tiingo has expanded into more DEXes and chains.

### Fund Fees

- Improved extraction methodology and error detection.

## 2022-06-30

### General

- New Year - New look! Check our our new API page.
- General performance improvements across several endpoints.
- Corrected API Documentation in several locations

### Crypto Synthetics

- We have added new products/pairs to be tracked in Crypto Synthetics.
- Reach out to [sales@tiingo.com](mailto:sales@tiingo.com) for more information.

### Crypto Feeds

- Tiingo has expanded into even more DEXes.

### IEX Feeds

- Added a new 1min bar OHLCV websocket option to the IEX Feed. Contact [sales@tiingo.com](mailto:sales@tiingo.com) for more details.
- Added a 15ms delay to the IEX Feed as required by the IEX Exchange to avoid passing a $500/month license fee to our customers. We believe the IEX Exchange has created the most fair licensing for equity data to date. 15ms should make no practical impact to customers since network latency between our servers in NJ/NY and customers will most likely be > 15ms alone given geographical separation.

## 2022-03-10

### General

- Added more types of bulk downloads for enterprise customers.
- Websocket feeds now allow additions to subscribed tickers within the same connection.
- Changed API usage tiers for new customers (old customers will remain at existing tiers).

### Crypto Synthetics - New Product

- We have launched a new CryptoCurrency benchmarking product: Crypto Synthetics.
- This product aggregates multiple cross pairs (e.g. ETH/AUD, ETH/USDC, ETH/BTC, etc.), converts them into a single quote pair (e.g. USD), VWAPs, and produces an aggregated synthetic feeds with microsecond latency - making it one of the fastest benchmarking products in existence.
- Reach out to [sales@tiingo.com](mailto:sales@tiingo.com) for more information.

### Crypto Feeds

- Tiingo has expanded into even more blockchains and DEXes (Solana, Terraswap, Arbitrum, and more)

### Fund Fees

- Improved fund fee extraction method to backfill older data.

### FX Feeds

- We have sourced new FX data, resulting in much tighter spreads.

### News API

- News Database has now hit 40,000,000+ articles.

### IEX Feeds

- Significantly improved speed of IEX endpoints.

## 2021-09-14

### Crypto Feeds

- Tiingo now supports 40+ Crypto exchanges.
- We've been expanding aggressively into Decentralized Exchanges (DEXes) - check them out in the [Crypto API](https://api.tiingo.com/docs/crypto/overview)!
- Now support cross-chain currencies (AVAX, Polygon, BSC).
- Added various new Crypto Centralized Exchanges (CEXes).

### News API

- News Database has now hit 35,000,000+ articles.

### IEX Feeds

- IEX Historical data is now queryable by PermaTicker for clients who have PermaTickers enabled.

## 2021-06-05

### General

- Increased queryable history for the intraday REST APIs (Crypto, FX, and IEX).
- Various bug and performance fixes.

### News API

- Improved crawling speed of our news crawler farm.

### FX Feeds

- Large scale improvements in the performance of the FX API.

### Mutual Fund and ETF Fee API

- Improved variability range of Fee parsing from Mutual Funds and ETFs, allowing for greater coverage.

## 2021-03-26

### EOD Composite Feeds

- Added a sort request field for the End-of-Day Historical endpoint.

### Tiingo Fundamentals

- Added a sort request field for Fundamentals Statements and Daily endpoints.

## 2021-03-23

### EOD Composite Feeds

- You can now query EOD price history by permaTicker for accounts that have permaTicker enabled.

### FX Feeds

- Improved frequency of FX price updates for the websocket and REST servers.
- Added back gold and silver to the Forex endpoints.
- FX was migrated to the same caching engine as Crypto and IEX for faster queries and performance.

### Crypto Feeds

- Added more crypto pairs for the feed as well as reworked infrastructure for more stable connections.

### Tiingo Fundamentals

- Fixed a calculation for certain tickers regarding market capitalization

## 2021-02-23

### Tiingo Fundamentals

- Fixed a date boundary bug when requesting fundamental data where an extra day was added to the response than what was requested. Check out the fundamentals documentation here: [Tiingo Fundamentals](https://www.tiingo.com/documentation/fundamentals).

## 2020-09-01

### EOD Composite Feeds

- Added an additional enterprise data source for EOD feeds for improved accuracy, coverage, and exchange listing changes.

## 2020-05-18

### Tiingo Fundamentals

- Significantly expanded the number of meta fields, including Sector & Industry data, locations, company websites, and more. Check out the fundamentals documentation here: [Tiingo Fundamentals](https://www.tiingo.com/documentation/fundamentals).

### IEX Feeds

- Added the Volume field for the historical intraday IEX data Endpoint (2.6.3). To get volume data, "volume" must be passed to the "columns" parameter. Check out the IEX documentation here: [IEX REST](https://www.tiingo.com/documentation/iex) to see an example.

## 2020-04-05

### Tiingo Fundamentals

- The DOW 30 is now available as a sample/evaluation dataset. 3 Years of data are available for each company in the DOW 30. Check out the fundamentals documentation here: [Tiingo Fundamentals](https://www.tiingo.com/documentation/fundamentals).

### FX Feeds

- Improved stability for the real-time Forex data.

## 2020-02-19

### New Product - Fundamentals

- We have added fundamentals to beta Visit the documentation here: [Tiingo Fundamentals](https://www.tiingo.com/documentation/fundamentals)

### Crypto Feeds

- A new intraday caching engine was deployed resulting in O(1) query times for historical intraday crypto data.

### IEX Feeds

- A new intraday caching engine was deployed resulting in O(1) query times for historical intraday IEX data.

## 2019-08-19

### New Product - Search

- We have added the ability to search for tickers in the Tiingo database. This is useful for building typeaheads. Visit the documentation here: [Tiingo Search](https://www.tiingo.com/documentation/utilities/search)

## 2019-07-28

### New Product - FX Feeds

- We have added institutional quality FX feeds into beta. Visit the documentation here: [Forex Feeds](https://www.tiingo.com/documentation/forex)

### IEX Feeds

- You can now query IEX data after hours data as well as the ability to force fill a timeseries (for cleaner charts) Check out the new historical data request parameters in the documentation here: [IEX Endpoint](https://www.tiingo.com/documentation/iex).

## 2019-05-09

### New Product - Company Descriptions

- We have added a new alternative data product that helps you parse through the delta of company descriptions and discover sentiment. Visit it here: [Company Descriptions](https://www.tiingo.com/products/company-descriptions)

### EOD Composite Feeds

- You can now query EOD data for crypto pairs using the [EOD Endpoint](https://www.tiingo.com/documentation/end-of-day).

## 2019-02-11

### Updated API Site

- Does something feel different? We've updated our entire API documentation to make it easier to find relevant information for you.
- [Documentation pages](https://www.tiingo.com/documentation/general/overview) have been made more complete, for example it now includes answers about our [Developer Program](https://www.tiingo.com/documentation/appendix/developers) and our [Symbology](https://www.tiingo.com/documentation/appendix/symbology) format.
- Product pages have been expanded to include more data about what is covered in each offering.

## 2018-08-26

### General

- We've added a status page to keep users informed of updates and maintenace. You can visit it here: [https://status.tiingo.com](https://status.tiingo.com/)
- Better error messages for when passing malformed dates to the endpoints

### Crypto Feeds

- We've added top-of-book quote updates to the websocket API. Check out our crypto API docs here: [https://api.tiingo.com/docs/crypto/overview](https://api.tiingo.com/docs/crypto/overview)

## 2018-03-24

### EOD Composite Feeds

- Improved FX conversion for ADRs that list their dividends in non-USD currencies.
- Faster update times for when exchanges send corrections.

## 2017-12-27

### Crypto Feeds

- Check out our new crypto API available here: [https://api.tiingo.com/docs/crypto/overview](https://api.tiingo.com/docs/crypto/overview)

### EOD Composite Feeds

- Faster loading and prioritization of asset prices.
- Expanded asset coverage.
- Now includes exchange corrections throughout the day.

## 2017-09-19

### General

- Adding a Integrations/Plugins page that contains programming packages, software, and integrations that allow you to use Tiingo's data. Check it out here: [Integrations](https://www.tiingo.com/documentation/general/changelog/https:/api.tiingo.com/resources/integrations).

### EOD Composite Feeds

- For international equities, the prices are reported in the country's currency rather than defaulting to USD.

## 2017-07-25

### General

- Server caching improved for faster performance.

### EOD Composite Feeds

- This endpoint can now return values as CSVs or JSON formats. Append &format=csv to the query to get CSV data. More information available in the docs.
- This endpoint can now return values resampled to weekly, monthly, and annual values. Simply append resampleFreq= followed by daily, weekly, monthly, or annually to see the resampled values. E.g. &resampleFreq=monthly for monthly returns. More details available in the docs.
- If no end date is provided, a default endDate of today is used. This means you can provide a start date and all history up until now will be provided.

### IEX API

- You can now query multiple tickers by sending a list or a comma-separated list of tickers to the request URL. e.g. https://api.tiingo.com/iex/spy,googl.

## 2017-07-14

### Auth token page

- You can now view your auth token at: [https://api.tiingo.com/account/token](https://api.tiingo.com/account/token).

## 2017-07-5

### News API

- News API has now been opening up to Power users! You may query up to a month of historical data and get the full features for any current data as our feeds update! Check it out here: [https://api.tiingo.com/docs/tiingo/news](https://api.tiingo.com/docs/tiingo/news). We can now offer this data because of our decision to go freemium.

### Pricing Page

- You can now upgrade on the API page by visiting: [https://api.tiingo.com/pricing](https://api.tiingo.com/pricing).

# General API Documentation Overview

> 출처: https://www.tiingo.com/documentation/general/overview · 생성: 2026-09-04 (tools/gendocs)

## 1.1 Overview

Tiingo's APIs are built to be performant, consistent, and also support extensive filters to speed up your development time.

Because of this, we hope you will take the time to read the documentation and learn about all of our features, but we also understand you may want to jump right in. So if you want to jump right in, click on any endpoint link, which are located on the sidebar to your left.

## 1.1.1 Introduction

The endpoints are broken up into two different types:

1. **REST Endpoints** which provide a RESTful interface to querying data, especially historical data for end-of-day data feeds.
2. **Websocket** which provide a websocket interface and are used to stream real-time data. If you are looking for access to a raw firehose of data, the websocket endpoint is how you get it.

To navigate and see how to query data using each of the above methods, use the sidebar on the left and choose an endpoint that interests you.

Additionally, there are a number of third-party packages that interface with our data and make it even easier to get up and running with our data. You can check them out here:

## 1.1.2 Authentication

In order to use the API, you must sign-up to create an account. All accounts are free and if you need higher usage limits, or you have a commercial use case, you can upgade to the Power and/or Commercial plan.

Once you create an account, your account with be assigned an authentication token. This token is used in place of your username & password throughout the API, so keep it safe like you would your password.

You can find your API token by clicking [here](https://www.tiingo.com/account/api/token), or to make it easy, your token will be shown below if you are logged in.

Your API Token is:

```
<TOKEN>
```

To see how to use your token to make requests, you can use the code examples throughout the documentation, or check out the section on connecting: [1.2.1 Connecting](https://www.tiingo.com/documentation/general/connecting).

## 1.1.3 Usage Limits

To keep the API affordable to all, each account is given generous rate-limits. We limit based on:

- Hourly Requests - Reset every hour.
- Daily Requests - Reset every day at midnight EST.
- Monthly Bandwidth - Reset the first of every month at midnight EST.

We do not rate limit to minute or second, so you are free to make your requests as you desire.

The basic, power, and commercial power plans offer different levels of rate limits. To see what these rate limits are, visit the [pricing page](https://www.tiingo.com/pricing).

If you need custom limits set you can E-mail our [sales team](mailto:sales@tiingo.com), who will give you a fair and reasonable price. Chances are, if you've been with us for some time and the request is reasonable, we will increase your limits at no charge. Do not hesitate to reach out, we are here for you.

## 1.1.4 Response Formats

For most of our REST endpoints, we allow you, the user, to choose which format the data data can be returned in. We support two different return formats:

- **JSON** - The data is returned in the JSON data structure. This format allows the most flexibility as we can append meta data as well as debugging data. The downside to this data type is that it requires more bandwidth, which means it may be slower since more data has to be downloaded and parsed by the client side.
- **CSV** - This is a "bare-bones" data return type that is often 4-5x faster than JSON. The data is returned in comma-separated-format, which is helpful when importing the data in spreadsheet programs like Excel.

To return data in a particular format, you may pass the **format** parameter, which can take both "json" and "csv".

When browsing the documentation, you will see at the top of the page which return formats are supported.

## 1.1.5 Symbology

Tiingo's symbol format uses dashes ("-") instead of periods (".") to denote share classes. Our API covers both common share classes and preferred share classes. For example Berkshire class A shares would be "BRK-A" and Simon Property Group's Preferred J series shares would be "SPG-P-J".

More details can be found on the [Symbology Appendix](https://www.tiingo.com/documentation/appendix/symbology) page and a full list of tickers can be found in [supported_tickers.zip](https://apimedia.tiingo.com/docs/tiingo/daily/supported_tickers.zip), which is updated daily.

## 1.1.6 Permitted Use of Our Data

For Basic and Power accounts, data is for internal and personal use only. You may not redistribute the data in any form.

For Commercial accounts, data is licensed for internal commercial usage. You may not redistribute the data in any form.

If you would like to redistribute the data for commercial or personal use, for example: a presentation, a proposal, a website or app, or any other usage case, please E-mail [sales@tiingo.com](mailto:sales@tiingo.com) and include the following:

- The use case for the data.
- A website link to your firm or academic affiliation.
- Whether your company is a start-up (less than 5 employees) or enterprise (5 or more employees).

All of our redistributable pricing is based on a flat rate, so it is predictable, simple, and easy. Additionally, redistributable licenses come with substantially higher usage limits to help you save storage costs and speed up loading of our data.

If you are a developer and are building software for your audience that requires users to submit their own Tiingo API token in order to use your software, and your software is not distributing our data, you do not need to contact us regarding licensing. Please read about our [developer program](https://www.tiingo.com/documentation/appendix/developers) if you are interesting in building software that integrates into Tiingo.

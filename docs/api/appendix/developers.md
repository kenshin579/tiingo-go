# Developer Program API Documentation

> 출처: https://www.tiingo.com/documentation/appendix/developers · 생성: 2026-09-04 (tools/gendocs)

## 5.1 Developer Program

This guide is meant for software developers who would like to integrate their financial software to use Tiingo's API as a data source, where their software requires users to supply their own Tiingo API Token. This requires users to register for their own Tiingo account.

This is not meant for developers digesting Tiingo data and then redistributing Tiingo's data on their own platform/app/software. If you would like to redistribute Tiingo's data without requiring your users to make their own Tiingo account, you will have to get a redistribution license. Please contact [sales@tiingo.com](mailto:sales@tiingo.com) to obtain a redistribution license.

## 5.1.1 Introduction

We love when developers integrate Tiingo as a data source into their software or software library. You do not need permission if your use case qualifies for our developer program, but please let us know about your software/library by E-mailing us at [developers@tiingo.com](mailto:developers@tiingo.com) so we can add it to [our growing list](https://www.tiingo.com/documentation/appendix/integrations). Additionally, we will add you to a developers mailing list so you may stay posted of the latest features and additions before it is open to the general public.

## 5.1.2 Registration into the Developer Program

To register for the developer program please E-mail [developers@tiingo.com](mailto:developers@tiingo.com) the following information:

- The name of your software and/or library.
- A link to your product's page or library page (github, bitbucket, and gitlab sites are perfectly fine).
- The description you would like us to use for your software on our [integrations page](https://www.tiingo.com/documentation/appendix/integrations).
- The best way to contact you (E-mail address, Phone, etc.). Only provide what you feel comfortable.
- **Optional:** A logo, image, text, and/or any other trademark ("marks") that we may use for marketing purposes across social media and Tiingo pages. This is entirely optional, but if you do send us such marks, you grant Tiingo Inc. a worldwide perpetual license to use such marks for the purposes of educating and marketing your software's integration into Tiingo.

By E-mailing us the above information, you agree to the Tiingo [Terms of Service](https://www.tiingo.com/tos), [Privacy Policy](https://www.tiingo.com/tos), and Developer Rules as outlined in section 5.1.4 ("Developer Program Rules") below.

## 5.1.3 Authorization Token

Each user must have their own API token. Once a user registers for Tiingo, guide them to this page to obtain their API Token: [Account - API Token](https://www.tiingo.com/account/api/token). You may then allow the user to paste their API token in your software and then use their API token to obtain Tiingo data.

We do not allow programmatic registration or logins due to security concerns so the API token is the only way to authenticate users.

## 5.1.4 Developer Program Rules

Our users trust us to keep their data secure and private. In order to keep compliant with the developer program, we require that you follow the rules below:

- Tiingo is a privacy conscious company. You may not track, record, document, distribute, or sell user API analytic meta data. You may only track which endpoints your users are using **if and only if** it is for general support, software debugging, or general internal improvement purposes. Additionally you agree to only track this data **if and only if** you make such tracking clear & explicit, the consent for tracking is opt-in, and the user may opt out at any time.
- The only form of authentication allows is via the API Authorization Token. Do not wrap registration/login forms as this present security concerns for users. You agree to never ask users for their Tiingo login password.
- You accept and acknowledge Tiingo does not endorse your software, library, or software created by you or your organization. You acknowledge that Tiingo is not liable for any data misuse by you or your end users and will notify us if you believe any of your users are abusing our APIs. You and your end users are bound by the Tiingo [Terms of Service](https://www.tiingo.com/tos) and [Privacy Policy](https://www.tiingo.com/privacy).
- Tiingo must be clearly attributed as the data source.
- If you violate the rules, we may either require you to remove the Tiingo integration or give you steps to fix the issue. Generally, we will work with you to help make your software compliant, unless the the error is malicious or not in good faith.

## 5.1.5 Our Commitment to You

We will try our hardest to keep the following commitments to our developers.

- If we create a breaking change, we will work to ensure we support the current functionality for at least 1 year (end-of-life deprecation schedule).
- We will communicate changes to the API on the changelog as well as via E-mail if you register for the developer program (E-mail [developers@tiingo.com](mailto:developers@tiingo.com) to register).
- We are here for your support and to help you integrate your software/library. Please reach out to us [developers@tiingo.com](mailto:developers@tiingo.com) for priority support.
- We will work to keep your end user's information confidential and provide support to your users where we can.

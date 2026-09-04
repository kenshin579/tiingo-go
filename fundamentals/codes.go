package fundamentals

// 이 파일은 /tiingo/fundamentals/definitions 응답에서 생성했다(2026-09-04, 85개).
// Tiingo 는 지표를 계속 추가하므로 목록이 최신이 아닐 수 있다. 상수가 없어도
// StatementData.Get 은 임의 문자열을 받으므로 동작에는 지장이 없다.

// 손익계산서(incomeStatement) 지표 22개
const (
	CodeConsolidatedIncome      = "consolidatedIncome"      // Consolidated Income ($)
	CodeCostRev                 = "costRev"                 // Cost of Revenue ($)
	CodeEbit                    = "ebit"                    // Earning Before Interest & Taxes EBIT ($)
	CodeEbitda                  = "ebitda"                  // EBITDA ($)
	CodeEbt                     = "ebt"                     // Earnings before tax ($)
	CodeEps                     = "eps"                     // Earnings Per Share ($)
	CodeEpsDil                  = "epsDil"                  // Earnings Per Share Diluted ($)
	CodeGrossProfit             = "grossProfit"             // Gross Profit ($)
	CodeIntexp                  = "intexp"                  // Interest Expense ($)
	CodeNetIncComStock          = "netIncComStock"          // Net Income Common Stock ($)
	CodeNetIncDiscOps           = "netIncDiscOps"           // Net Income from Discontinued Operations ($)
	CodeNetinc                  = "netinc"                  // Net Income ($)
	CodeNonControllingInterests = "nonControllingInterests" // Net Income to Non-Controlling Interests ($)
	CodeOpex                    = "opex"                    // Operating Expenses ($)
	CodeOpinc                   = "opinc"                   // Operating Income ($)
	CodePrefDVDs                = "prefDVDs"                // Preferred Dividends Income Statement Impact ($)
	CodeRevenue                 = "revenue"                 // Revenue ($)
	CodeRnd                     = "rnd"                     // Research & Development ($)
	CodeSga                     = "sga"                     // Selling, General & Administrative ($)
	CodeShareswa                = "shareswa"                // Weighted Average Shares
	CodeShareswaDil             = "shareswaDil"             // Weighted Average Shares Diluted
	CodeTaxExp                  = "taxExp"                  // Tax Expense ($)
)

// 재무상태표(balanceSheet) 지표 26개
const (
	CodeAccoci                = "accoci"                // Accumulated Other Comprehensive Income ($)
	CodeAcctPay               = "acctPay"               // Accounts Payable ($)
	CodeAcctRec               = "acctRec"               // Accounts Receivable ($)
	CodeAssetsCurrent         = "assetsCurrent"         // Current Assets ($)
	CodeAssetsNonCurrent      = "assetsNonCurrent"      // Other Assets ($)
	CodeCashAndEq             = "cashAndEq"             // Cash and Equivalents ($)
	CodeDebt                  = "debt"                  // Total Debt ($)
	CodeDebtCurrent           = "debtCurrent"           // Current Debt ($)
	CodeDebtNonCurrent        = "debtNonCurrent"        // Non-Current Debt ($)
	CodeDeferredRev           = "deferredRev"           // Deferred Revenue ($)
	CodeDeposits              = "deposits"              // Deposits ($)
	CodeEquity                = "equity"                // Shareholders Equity ($)
	CodeIntangibles           = "intangibles"           // Intangible Assets ($)
	CodeInventory             = "inventory"             // Inventory ($)
	CodeInvestments           = "investments"           // Investments ($)
	CodeInvestmentsCurrent    = "investmentsCurrent"    // Current Investments ($)
	CodeInvestmentsNonCurrent = "investmentsNonCurrent" // Non-Current Investments ($)
	CodeLiabilitiesCurrent    = "liabilitiesCurrent"    // Current Liabilities ($)
	CodeLiabilitiesNonCurrent = "liabilitiesNonCurrent" // Other Liabilities ($)
	CodePpeq                  = "ppeq"                  // Property, Plant & Equipment ($)
	CodeRetainedEarnings      = "retainedEarnings"      // Accumulated Retained Earnings or Deficit ($)
	CodeSharesBasic           = "sharesBasic"           // Shares Outstanding
	CodeTaxAssets             = "taxAssets"             // Tax Assets ($)
	CodeTaxLiabilities        = "taxLiabilities"        // Tax Liabilities ($)
	CodeTotalAssets           = "totalAssets"           // Total Assets ($)
	CodeTotalLiabilities      = "totalLiabilities"      // Total Liabilities ($)
)

// 현금흐름표(cashFlow) 지표 14개
const (
	CodeBusinessAcqDisposals    = "businessAcqDisposals"    // Business Acquisitions & Disposals ($)
	CodeCapex                   = "capex"                   // Capital Expenditure ($)
	CodeDepamor                 = "depamor"                 // Depreciation, Amortization & Accretion ($)
	CodeFreeCashFlow            = "freeCashFlow"            // Free Cash Flow ($)
	CodeInvestmentsAcqDisposals = "investmentsAcqDisposals" // Investment Acquisitions & Disposals ($)
	CodeIssrepayDebt            = "issrepayDebt"            // Issuance or Repayment of Debt Securities ($)
	CodeIssrepayEquity          = "issrepayEquity"          // Issuance or Repayment of Equity ($)
	CodeNcf                     = "ncf"                     // Net Cash Flow to Change in Cash & Cash Equivalents ($)
	CodeNcff                    = "ncff"                    // Net Cash Flow from Financing ($)
	CodeNcfi                    = "ncfi"                    // Net Cash Flow from Investing ($)
	CodeNcfo                    = "ncfo"                    // Net Cash Flow from Operations ($)
	CodeNcfx                    = "ncfx"                    // Effect of Exchange Rate Changes on Cash ($)
	CodePayDiv                  = "payDiv"                  // Payment of Dividends & Other Cash Distributions ($)
	CodeSbcomp                  = "sbcomp"                  // Shared-based Compensation ($)
)

// 개요·비율(overview) 지표 23개
const (
	CodeAssetTurnover      = "assetTurnover"      // Asset Turnover
	CodeBookVal            = "bookVal"            // Book Value ($)
	CodeBvps               = "bvps"               // Book Value Per Share ($)
	CodeCurrentRatio       = "currentRatio"       // Current Ratio
	CodeDebtEquity         = "debtEquity"         // Debt to Equity Ratio
	CodeEnterpriseVal      = "enterpriseVal"      // Enterprise Value ($)
	CodeEpsQoQ             = "epsQoQ"             // Earnings Per Share QoQ Growth (%)
	CodeFxRate             = "fxRate"             // FX Rate
	CodeGrossMargin        = "grossMargin"        // Gross Margin (%)
	CodeLongTermDebtEquity = "longTermDebtEquity" // Long-term Debt to Equity
	CodeMarketCap          = "marketCap"          // Market Capitalization ($)
	CodeNetMargin          = "netMargin"          // Net Margin (%)
	CodeOpMargin           = "opMargin"           // Operating Margin (%)
	CodePbRatio            = "pbRatio"            // Price to Book Ratio
	CodePeRatio            = "peRatio"            // Price to Earnings Ratio
	CodePiotroskiFScore    = "piotroskiFScore"    // Piotroski F-Score
	CodeProfitMargin       = "profitMargin"       // Profit Margin (%)
	CodeRevenueQoQ         = "revenueQoQ"         // Revenue QoQ Growth (%)
	CodeRoa                = "roa"                // Return on Assets ROA (%)
	CodeRoe                = "roe"                // Return on Equity ROE (%)
	CodeRps                = "rps"                // Revenue Per Share ($)
	CodeShareFactor        = "shareFactor"        // Share Factor
	CodeTrailingPEG1Y      = "trailingPEG1Y"      // PEG Ratio
)

// AllCodes 는 위 상수 전체다. 정의 목록과의 동기화 검사에 쓴다.
var AllCodes = []string{
	CodeConsolidatedIncome,
	CodeCostRev,
	CodeEbit,
	CodeEbitda,
	CodeEbt,
	CodeEps,
	CodeEpsDil,
	CodeGrossProfit,
	CodeIntexp,
	CodeNetIncComStock,
	CodeNetIncDiscOps,
	CodeNetinc,
	CodeNonControllingInterests,
	CodeOpex,
	CodeOpinc,
	CodePrefDVDs,
	CodeRevenue,
	CodeRnd,
	CodeSga,
	CodeShareswa,
	CodeShareswaDil,
	CodeTaxExp,
	CodeAccoci,
	CodeAcctPay,
	CodeAcctRec,
	CodeAssetsCurrent,
	CodeAssetsNonCurrent,
	CodeCashAndEq,
	CodeDebt,
	CodeDebtCurrent,
	CodeDebtNonCurrent,
	CodeDeferredRev,
	CodeDeposits,
	CodeEquity,
	CodeIntangibles,
	CodeInventory,
	CodeInvestments,
	CodeInvestmentsCurrent,
	CodeInvestmentsNonCurrent,
	CodeLiabilitiesCurrent,
	CodeLiabilitiesNonCurrent,
	CodePpeq,
	CodeRetainedEarnings,
	CodeSharesBasic,
	CodeTaxAssets,
	CodeTaxLiabilities,
	CodeTotalAssets,
	CodeTotalLiabilities,
	CodeBusinessAcqDisposals,
	CodeCapex,
	CodeDepamor,
	CodeFreeCashFlow,
	CodeInvestmentsAcqDisposals,
	CodeIssrepayDebt,
	CodeIssrepayEquity,
	CodeNcf,
	CodeNcff,
	CodeNcfi,
	CodeNcfo,
	CodeNcfx,
	CodePayDiv,
	CodeSbcomp,
	CodeAssetTurnover,
	CodeBookVal,
	CodeBvps,
	CodeCurrentRatio,
	CodeDebtEquity,
	CodeEnterpriseVal,
	CodeEpsQoQ,
	CodeFxRate,
	CodeGrossMargin,
	CodeLongTermDebtEquity,
	CodeMarketCap,
	CodeNetMargin,
	CodeOpMargin,
	CodePbRatio,
	CodePeRatio,
	CodePiotroskiFScore,
	CodeProfitMargin,
	CodeRevenueQoQ,
	CodeRoa,
	CodeRoe,
	CodeRps,
	CodeShareFactor,
	CodeTrailingPEG1Y,
}

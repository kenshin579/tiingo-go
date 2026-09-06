//go:build integration

package tiingo_test

import (
	"context"
	"os"
	"testing"
	"time"

	tiingo "github.com/kenshin579/tiingo-go"
	"github.com/kenshin579/tiingo-go/corporateactions"
	"github.com/kenshin579/tiingo-go/crypto"
	"github.com/kenshin579/tiingo-go/eod"
	"github.com/kenshin579/tiingo-go/equity"
	"github.com/kenshin579/tiingo-go/forex"
	"github.com/kenshin579/tiingo-go/fundamentals"
	"github.com/kenshin579/tiingo-go/iex"
	"github.com/kenshin579/tiingo-go/search"
	"github.com/kenshin579/tiingo-go/stream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 실행: TIINGO_API_KEY=... go test -tags integration ./...
func newClient(t *testing.T) *tiingo.Client {
	t.Helper()
	if os.Getenv(tiingo.APIKeyEnv) == "" {
		t.Skip(tiingo.APIKeyEnv + " not set")
	}
	c, err := tiingo.NewClientFromEnv()
	require.NoError(t, err)
	return c
}

func TestIntegration_Meta(t *testing.T) {
	c := newClient(t)
	m, err := c.EOD.Meta(context.Background(), "AAPL")
	require.NoError(t, err)
	assert.Equal(t, "AAPL", m.Ticker)
	assert.NotEmpty(t, m.Name)
	assert.Equal(t, "NASDAQ", m.ExchangeCode)
	require.NotNil(t, m.StartDate)
	assert.Equal(t, "1980-12-12", m.StartDate.String())
}

func TestIntegration_LatestPrice(t *testing.T) {
	c := newClient(t)
	p, err := c.EOD.LatestPrice(context.Background(), "AAPL")
	require.NoError(t, err)
	assert.False(t, p.Date.IsZero())
	assert.Greater(t, p.Close, 0.0)
	assert.Greater(t, p.AdjClose, 0.0)
}

func TestIntegration_HistoricalPrices(t *testing.T) {
	c := newClient(t)
	ps, err := c.EOD.HistoricalPrices(context.Background(), "AAPL", &eod.PriceOptions{
		StartDate:    time.Date(2019, 1, 2, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2019, 1, 7, 0, 0, 0, 0, time.UTC),
		ResampleFreq: eod.ResampleDaily,
	})
	require.NoError(t, err)
	require.NotEmpty(t, ps)
	assert.Equal(t, "2019-01-02", ps[0].Date.String())
	assert.InDelta(t, 157.92, ps[0].Close, 0.01)
}

func TestIntegration_UnknownTicker(t *testing.T) {
	c := newClient(t)
	_, err := c.EOD.Meta(context.Background(), "NOSUCHTICKERXYZ")
	assert.Error(t, err)
}

func TestIntegration_FundamentalsDefinitions(t *testing.T) {
	c := newClient(t)
	ds, err := c.Fundamentals.Definitions(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, ds)

	known := map[string]bool{}
	for _, code := range fundamentals.AllCodes {
		known[code] = true
	}
	var missing []string
	for _, d := range ds {
		if !known[d.DataCode] {
			missing = append(missing, d.DataCode)
		}
	}
	assert.Emptyf(t, missing, "codes.go 갱신 필요 — 새 dataCode: %v", missing)
}

func TestIntegration_FundamentalsMeta(t *testing.T) {
	c := newClient(t)
	ms, err := c.Fundamentals.Meta(context.Background(), "AAPL")
	require.NoError(t, err)
	require.Len(t, ms, 1)
	assert.Equal(t, "Apple Inc", ms[0].Name)
	assert.False(t, ms[0].StatementLastUpdated.IsZero())
}

func TestIntegration_FundamentalsStatements(t *testing.T) {
	c := newClient(t)
	ss, err := c.Fundamentals.Statements(context.Background(), "AAPL", &fundamentals.StatementOptions{
		StartDate: time.Now().AddDate(-1, 0, 0),
		Sort:      "-date",
	})
	require.NoError(t, err)
	require.NotEmpty(t, ss)
	rev, ok := ss[0].StatementData.Get(fundamentals.CodeRevenue)
	assert.True(t, ok)
	assert.Greater(t, rev, 0.0)
}

func TestIntegration_FundamentalsDaily(t *testing.T) {
	c := newClient(t)
	ds, err := c.Fundamentals.Daily(context.Background(), "AAPL", &fundamentals.DailyOptions{
		StartDate: time.Now().AddDate(0, 0, -14),
	})
	require.NoError(t, err)
	require.NotEmpty(t, ds)
	assert.Greater(t, ds[0].MarketCap, 0.0)
}

func TestIntegration_CryptoMeta(t *testing.T) {
	c := newClient(t)
	ms, err := c.Crypto.Meta(context.Background(), "btcusd", "ethusd")
	require.NoError(t, err)
	require.Len(t, ms, 2)
	assert.Equal(t, "btcusd", ms[0].Ticker)
	assert.Equal(t, "btc", ms[0].BaseCurrency)
}

func TestIntegration_CryptoPrices(t *testing.T) {
	c := newClient(t)
	ss, err := c.Crypto.Prices(context.Background(), []string{"btcusd"}, &crypto.PriceOptions{
		StartDate:    time.Now().AddDate(0, 0, -3),
		ResampleFreq: crypto.Resample1Day,
	})
	require.NoError(t, err)
	require.NotEmpty(t, ss)
	require.NotEmpty(t, ss[0].PriceData)
	assert.Greater(t, ss[0].PriceData[0].Close, 0.0)
}

// 1분봉은 시각이 보존돼야 한다 — types.Time 사용의 실호출 확인.
func TestIntegration_CryptoIntraday(t *testing.T) {
	c := newClient(t)
	s, err := c.Crypto.PricesFor(context.Background(), "btcusd", &crypto.PriceOptions{
		ResampleFreq: crypto.Resample1Min,
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(s.PriceData), 2)
	d0, d1 := s.PriceData[0].Date.Time, s.PriceData[1].Date.Time
	assert.Equal(t, time.Minute, d1.Sub(d0))
}

func TestIntegration_CryptoTopOfBook(t *testing.T) {
	c := newClient(t)
	b, err := c.Crypto.TopOfBookFor(context.Background(), "btcusd", nil)
	require.NoError(t, err)
	require.NotEmpty(t, b.TopOfBookData)
	d := b.TopOfBookData[0]
	assert.Greater(t, d.BidPrice, 0.0)
	assert.Greater(t, d.AskPrice, 0.0)
	assert.False(t, d.QuoteTimestamp.IsZero())
}

func TestIntegration_CryptoUnknownTicker(t *testing.T) {
	c := newClient(t)
	_, err := c.Crypto.PricesFor(context.Background(), "nosuchpairxyz", nil)
	assert.Error(t, err, "없는 페어는 에러(빈 배열이면 ErrNotFound)")
}

func TestIntegration_ForexTopOfBook(t *testing.T) {
	c := newClient(t)
	qs, err := c.Forex.TopOfBook(context.Background(), []string{"eurusd", "usdjpy"})
	require.NoError(t, err)
	require.Len(t, qs, 2)
	assert.Equal(t, "eurusd", qs[0].Ticker)
	assert.Greater(t, qs[0].MidPrice, 0.0)
	assert.False(t, qs[0].QuoteTimestamp.IsZero())
}

// 주말에도 실패하지 않도록 최근 7일을 조회하고, 빈 응답을 허용한다.
func TestIntegration_ForexPrices(t *testing.T) {
	c := newClient(t)
	ps, err := c.Forex.Prices(context.Background(), []string{"eurusd"}, &forex.PriceOptions{
		StartDate:    time.Now().AddDate(0, 0, -7),
		ResampleFreq: forex.Resample1Day,
	})
	require.NoError(t, err)
	for _, p := range ps {
		assert.Equal(t, "eurusd", p.Ticker)
		assert.Greater(t, p.Close, 0.0)
	}
}

func TestIntegration_ForexMultiTickerPath(t *testing.T) {
	c := newClient(t)
	ps, err := c.Forex.Prices(context.Background(), []string{"eurusd", "usdjpy"}, &forex.PriceOptions{
		StartDate:    time.Now().AddDate(0, 0, -7),
		ResampleFreq: forex.Resample1Day,
	})
	require.NoError(t, err)
	if len(ps) == 0 {
		t.Skip("조회 구간이 전부 휴장일")
	}
	seen := map[string]bool{}
	for _, p := range ps {
		seen[p.Ticker] = true
	}
	assert.Len(t, seen, 2, "복수 티커가 경로로 전달돼 두 통화쌍이 와야 한다")
}

// 없는 통화쌍은 에러가 아니라 빈 슬라이스다(주말과 구분되지 않는다).
func TestIntegration_ForexUnknownTicker(t *testing.T) {
	c := newClient(t)
	ps, err := c.Forex.Prices(context.Background(), []string{"nosuchpair"}, nil)
	require.NoError(t, err)
	assert.Empty(t, ps)
}

// 장 마감·개장 어느 때 돌려도 깨지지 않도록 nil 여부는 단정하지 않는다.
func TestIntegration_IEXQuotes(t *testing.T) {
	c := newClient(t)
	qs, err := c.IEX.Quotes(context.Background(), []string{"AAPL", "MSFT"})
	require.NoError(t, err)
	require.Len(t, qs, 2)
	for _, q := range qs {
		assert.NotEmpty(t, q.Ticker)
		assert.False(t, q.Timestamp.IsZero())
		assert.Greater(t, q.TngoLast, 0.0)
		assert.Greater(t, q.PrevClose, 0.0)
		if q.BidPrice != nil {
			assert.Greater(t, *q.BidPrice, 0.0, "호가가 있으면 양수여야 한다")
		}
	}
}

// 없는 티커는 응답에서 빠진다.
func TestIntegration_IEXUnknownTickerOmitted(t *testing.T) {
	c := newClient(t)
	qs, err := c.IEX.Quotes(context.Background(), []string{"AAPL", "NOSUCHTICKERXYZ"})
	require.NoError(t, err)
	assert.Len(t, qs, 1, "없는 티커는 빠지고 AAPL 만 온다")
}

func TestIntegration_IEXPrices(t *testing.T) {
	c := newClient(t)
	ps, err := c.IEX.Prices(context.Background(), "AAPL", &iex.PriceOptions{
		StartDate:    time.Now().AddDate(0, 0, -5),
		ResampleFreq: iex.Resample1Hour,
	})
	require.NoError(t, err)
	for _, p := range ps {
		assert.False(t, p.Date.IsZero())
		assert.Greater(t, p.Close, 0.0)
	}
}

// columns 에 volume 을 넣어야 거래량이 온다.
func TestIntegration_IEXPricesVolume(t *testing.T) {
	c := newClient(t)
	ps, err := c.IEX.Prices(context.Background(), "AAPL", &iex.PriceOptions{
		StartDate:    time.Now().AddDate(0, 0, -5),
		ResampleFreq: iex.Resample1Hour,
		Columns:      []string{"open", "high", "low", "close", "volume"},
	})
	require.NoError(t, err)
	if len(ps) == 0 {
		t.Skip("조회 구간이 전부 휴장일")
	}
	var withVolume int
	for _, p := range ps {
		if p.Volume > 0 {
			withVolume++
		}
	}
	assert.Positive(t, withVolume, "columns 에 volume 을 넣으면 일부 구간은 거래량이 채워진다")
}

func TestIntegration_Search(t *testing.T) {
	c := newClient(t)
	rs, err := c.Search.Search(context.Background(), "apple", nil)
	require.NoError(t, err)
	require.NotEmpty(t, rs)
	for _, r := range rs {
		assert.NotEmpty(t, r.Ticker)
		assert.NotEmpty(t, r.PermaTicker)
		assert.NotEqual(t, search.FIGI("nan"), r.OpenFIGIComposite, "결손 값은 빈 문자열로 정규화된다")
	}
}

// ISIN 으로 자산 하나를 지목한다. 문서 표에 없는 파라미터라 실 API 로 확인해 둔다.
func TestIntegration_SearchByISIN(t *testing.T) {
	c := newClient(t)
	rs, err := c.Search.SearchByISIN(context.Background(), "US0378331005", nil)
	require.NoError(t, err)
	require.Len(t, rs, 1)
	assert.Equal(t, "AAPL", rs[0].Ticker)
	assert.Equal(t, "US", rs[0].CountryCode)
}

func TestIntegration_SearchExactTickerMatch(t *testing.T) {
	c := newClient(t)
	rs, err := c.Search.Search(context.Background(), "aapl", &search.SearchOptions{ExactTickerMatch: true})
	require.NoError(t, err)
	require.NotEmpty(t, rs)
	for _, r := range rs {
		assert.Equal(t, "AAPL", r.Ticker, "정확히 일치하는 티커만 온다")
	}
}

// 없는 자산은 에러가 아니라 빈 슬라이스다.
func TestIntegration_SearchEmptyResult(t *testing.T) {
	c := newClient(t)
	rs, err := c.Search.Search(context.Background(), "zzzznosuchassetxyz", nil)
	require.NoError(t, err)
	assert.Empty(t, rs)
}

// 장 마감·주말에도 깨지지 않도록 Lq* 의 nil 여부는 단정하지 않는다.
func TestIntegration_EquitySnapshots(t *testing.T) {
	c := newClient(t)
	ss, err := c.Equity.Snapshots(context.Background(), []string{"AAPL", "SPY"})
	require.NoError(t, err)
	require.Len(t, ss, 2)
	for _, s := range ss {
		assert.NotEmpty(t, s.Ticker)
		assert.False(t, s.Timestamp.IsZero())
		assert.Greater(t, s.TngoLast, 0.0)
		assert.Greater(t, s.PrevClose, 0.0)
		assert.Greater(t, s.LqRefPrice, 0.0)
		if s.LqSpread != nil {
			assert.Greater(t, *s.LqSpread, 0.0, "유동성 지표가 있으면 양수여야 한다")
		}
	}
}

// 없는 티커는 응답에서 빠진다.
func TestIntegration_EquityUnknownTickerOmitted(t *testing.T) {
	c := newClient(t)
	ss, err := c.Equity.Snapshots(context.Background(), []string{"AAPL", "NOSUCHTICKERXYZ"})
	require.NoError(t, err)
	assert.Len(t, ss, 1, "없는 티커는 빠지고 AAPL 만 온다")
}

// 전 종목 조회는 약 5MB 라 한 건만 돌린다. 포인터로 둔 근거를 실 API 로 확인한다.
func TestIntegration_EquityAllSnapshots(t *testing.T) {
	c := newClient(t)
	ss, err := c.Equity.AllSnapshots(context.Background())
	require.NoError(t, err)
	assert.Greater(t, len(ss), 1000, "전 종목이라 수천 건이 온다")

	var withLq int
	for _, s := range ss {
		if s.LqSpread != nil {
			withLq++
		}
	}
	// 비어 있는 행이 있다는 것만 단정한다 — 포인터로 둔 이유가 이것이다.
	// 채워진 행의 개수는 장 마감 후 경과 시간에 따라 달라진다. 실측상 금요일 종가 직후
	// (토 오전)에는 44% 가 채워져 있었지만 하루 지난 일요일에는 0 이었다.
	assert.Less(t, withLq, len(ss), "유동성 지표가 비어 있는 행이 있다")
	t.Logf("유동성 지표가 채워진 행: %d / %d", withLq, len(ss))
}

func TestIntegration_EquityPrices(t *testing.T) {
	c := newClient(t)
	ps, err := c.Equity.Prices(context.Background(), "AAPL", &equity.PriceOptions{
		StartDate:    time.Now().AddDate(0, 0, -5),
		ResampleFreq: equity.Resample1Hour,
		Columns:      []string{"open", "high", "low", "close", "volume"},
	})
	require.NoError(t, err)
	for _, p := range ps {
		assert.False(t, p.Date.IsZero())
		assert.Greater(t, p.Close, 0.0)
	}
}

func TestIntegration_DistributionYield(t *testing.T) {
	c := newClient(t)
	ys, err := c.CorporateActions.DistributionYield(context.Background(), "AAPL",
		&corporateactions.YieldOptions{StartDate: time.Now().AddDate(0, -1, 0)})
	require.NoError(t, err)
	require.NotEmpty(t, ys)
	for _, y := range ys {
		assert.False(t, y.Date.IsZero())
		assert.GreaterOrEqual(t, y.TrailingDiv1Y, 0.0, "수익률은 음수가 될 수 없다")
	}
}

// 기간을 좁히면 실제로 줄어든다 — 문서 표에 없는 파라미터라 실 API 로 확인해 둔다.
func TestIntegration_DistributionYieldDateRange(t *testing.T) {
	c := newClient(t)
	narrow, err := c.CorporateActions.DistributionYield(context.Background(), "AAPL",
		&corporateactions.YieldOptions{
			StartDate: time.Now().AddDate(0, 0, -10),
			EndDate:   time.Now().AddDate(0, 0, -5),
		})
	require.NoError(t, err)
	assert.Less(t, len(narrow), 10, "닷새 구간이라 거래일 수만큼만 온다")
}

// 없는 티커는 404 가 아니라 빈 슬라이스다.
func TestIntegration_DistributionYieldUnknownTicker(t *testing.T) {
	c := newClient(t)
	ys, err := c.CorporateActions.DistributionYield(context.Background(), "NOSUCHTICKERXYZ", nil)
	require.NoError(t, err)
	assert.Empty(t, ys)
}

// crypto 는 24시간 거래되므로 언제 돌려도 메시지가 온다.
// mid = (bid+ask)/2 검산으로 배열 매핑이 살아 있는지 매번 다시 확인한다.
func TestIntegration_StreamCrypto(t *testing.T) {
	c := newClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	s, err := c.Stream.Crypto(ctx, &stream.CryptoOptions{
		Tickers:   []string{"btcusd", "ethusd"},
		Threshold: stream.CryptoTradesAndQuotesLevel,
	})
	require.NoError(t, err)
	defer s.Close()

	var got int
	for msg := range s.Messages() {
		switch m := msg.(type) {
		case stream.CryptoTrade:
			assert.NotEmpty(t, m.Ticker)
			assert.NotEmpty(t, m.Exchange)
			assert.Greater(t, m.LastPrice, 0.0)
			assert.False(t, m.Date.IsZero())
		case stream.CryptoQuote:
			assert.NotEmpty(t, m.Ticker)
			assert.Greater(t, m.BidPrice, 0.0)
			assert.Greater(t, m.AskPrice, 0.0)
			assert.InDelta(t, (m.BidPrice+m.AskPrice)/2, m.MidPrice, 1e-6,
				"매핑이 맞으면 mid 가 검산된다")
		default:
			t.Fatalf("예상 밖 메시지 타입 %T", msg)
		}
		got++
		if got >= 5 {
			break
		}
	}
	assert.GreaterOrEqual(t, got, 1, "20초 안에 최소 한 건은 온다")
	assert.Equal(t, int64(0), s.Reconnects(), "20초 안에 재연결이 일어나면 안 된다")
}

// 구독 확인 id 가 온다 — 인증이 실제로 통했다는 증거다.
func TestIntegration_StreamSubscriptionID(t *testing.T) {
	c := newClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	s, err := c.Stream.Crypto(ctx, &stream.CryptoOptions{Tickers: []string{"btcusd"}})
	require.NoError(t, err)
	defer s.Close()

	assert.Eventually(t, func() bool { return s.SubscriptionID() != 0 },
		10*time.Second, 100*time.Millisecond)
}

// BOATS 는 계정 권한이 없어 E 프레임(403)이 온다. 그 경로가 에러로 잘 흐르고 재연결하지 않는지 확인한다.
// 이 테스트가 실패하면 계정에 BOATS 권한이 생긴 것일 수 있다 — 그렇다면 좋은 소식이니 그대로 보고한다.
func TestIntegration_StreamBOATSForbidden(t *testing.T) {
	c := newClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	s, err := c.Stream.BOATS(ctx, nil)
	require.NoError(t, err)
	defer s.Close()

	for range s.Messages() { //nolint:revive // 채널이 닫힐 때까지 비운다
	}
	require.Error(t, s.Err(), "권한이 없으므로 에러로 끝난다")
	assert.Contains(t, s.Err().Error(), "403")
	assert.Equal(t, int64(0), s.Reconnects(), "권한 거부 뒤에는 재연결하지 않는다")
}

package stream

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 실측 캡처로 crypto 매핑을 고정한다.
func TestDecodeCrypto_Trade(t *testing.T) {
	raw := json.RawMessage(`["T","btcusd","2026-09-06T03:55:54.366000+00:00","bitfinex",0.0002,80132.00000000001]`)
	m, err := decodeCrypto(raw)
	require.NoError(t, err)

	tr, ok := m.(CryptoTrade)
	require.True(t, ok)
	assert.Equal(t, "btcusd", tr.Ticker)
	assert.Equal(t, "bitfinex", tr.Exchange)
	assert.InDelta(t, 0.0002, tr.LastSize, 1e-12)
	assert.InDelta(t, 80132.0, tr.LastPrice, 1e-6)
	assert.False(t, tr.Date.IsZero())
}

// 실측 호가. mid = (bid+ask)/2 로 매핑이 맞는지 검산한다.
func TestDecodeCrypto_Quote(t *testing.T) {
	raw := json.RawMessage(`["Q","btcusd","2026-09-06T03:56:52.172908+00:00","gdax",6.517e-5,80046.4,80046.405,0.63361423,80046.41]`)
	m, err := decodeCrypto(raw)
	require.NoError(t, err)

	q, ok := m.(CryptoQuote)
	require.True(t, ok)
	assert.Equal(t, "gdax", q.Exchange)
	assert.InDelta(t, 6.517e-5, q.BidSize, 1e-12)
	assert.InDelta(t, 80046.4, q.BidPrice, 1e-6)
	assert.InDelta(t, 0.63361423, q.AskSize, 1e-9)
	assert.InDelta(t, 80046.41, q.AskPrice, 1e-6)
	assert.InDelta(t, (q.BidPrice+q.AskPrice)/2, q.MidPrice, 1e-6, "매핑이 맞으면 mid 가 검산된다")
}

// 실측 캡처 전체가 디코딩되고 모든 호가가 mid 검산을 통과해야 한다.
func TestDecodeCrypto_LiveCapture(t *testing.T) {
	raw, err := os.ReadFile("testdata/crypto_live.jsonl")
	require.NoError(t, err)

	var trades, quotes int
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		e, err := parseEnvelope([]byte(line))
		require.NoError(t, err)
		if e.MessageType != msgData {
			continue
		}
		m, err := decodeCrypto(e.Data)
		require.NoError(t, err, line)
		switch v := m.(type) {
		case CryptoTrade:
			assert.Greater(t, v.LastPrice, 0.0)
			trades++
		case CryptoQuote:
			assert.InDelta(t, (v.BidPrice+v.AskPrice)/2, v.MidPrice, 1e-6)
			quotes++
		default:
			t.Fatalf("예상 밖 타입 %T", m)
		}
	}
	assert.Positive(t, trades)
	assert.Positive(t, quotes)
}

// Forex 문서의 인덱스 표는 askPrice=6 이라 적지만 문서 예시의 산술은 askPrice=7 임을 증명한다.
// 실제 문서 예시(eurnok)를 그대로 써서 못 박는다 — 값을 지어내면 검산이 무의미해진다.
func TestDecodeForex_Quote_DocIndexTableIsWrong(t *testing.T) {
	// ["Q", ticker, date, bidSize, bidPrice, midPrice, askSize, askPrice]
	raw := json.RawMessage(`["Q","eurnok","2019-07-05T15:49:15.157000+00:00",5000000.0,9.6764,9.678135,5000000.0,9.67987]`)
	m, err := decodeForex(raw)
	require.NoError(t, err)

	q, ok := m.(ForexQuote)
	require.True(t, ok)
	assert.Equal(t, "eurnok", q.Ticker)
	assert.InDelta(t, 5000000.0, q.BidSize, 1e-9)
	assert.InDelta(t, 9.6764, q.BidPrice, 1e-9)
	assert.InDelta(t, 5000000.0, q.AskSize, 1e-9, "인덱스 6 은 수량이다(표는 가격이라 적었다)")
	assert.InDelta(t, 9.67987, q.AskPrice, 1e-9, "인덱스 7 이 가격이다(표는 수량이라 적었다)")
	assert.InDelta(t, (q.BidPrice+q.AskPrice)/2, q.MidPrice, 1e-9,
		"매핑이 맞으면 mid 가 검산된다 — 이 검산이 표를 반증한 근거다")
}

// IEX 기준가는 종류 문자가 없고 3원소다.
func TestDecodeIEX_ReferencePrice(t *testing.T) {
	raw := json.RawMessage(`["2026-09-06T13:30:00.000000000+00:00","AAPL",320.08]`)
	m, err := decodeIEX(raw)
	require.NoError(t, err)

	p, ok := m.(IEXReferencePrice)
	require.True(t, ok)
	assert.Equal(t, "AAPL", p.Ticker)
	assert.InDelta(t, 320.08, p.ReferencePrice, 1e-9)
	assert.False(t, p.Date.IsZero())
}

// Equity 는 두 메시지 모두 종류 문자가 없어 길이로 구분한다.
func TestDecodeEquity_ByLength(t *testing.T) {
	ref := json.RawMessage(`["2026-09-06T13:30:00+00:00","AAPL",320.08]`)
	m, err := decodeEquity(ref)
	require.NoError(t, err)
	_, ok := m.(EquityReferencePrice)
	assert.True(t, ok, "3원소는 기준가다")

	// 문서 예시 그대로: [date, ticker, spread, bidSize, bidPrice, refPrice, askPrice, askSize]
	lq := json.RawMessage(`["2019-01-30T13:33:45.383129126-05:00","aapl",0.0011,100,162.3,162.33,162.36,100]`)
	m, err = decodeEquity(lq)
	require.NoError(t, err)
	l, ok := m.(EquityLiquidity)
	require.True(t, ok, "8원소는 유동성이다")
	assert.InDelta(t, 0.0011, l.Spread, 1e-9)
	assert.InDelta(t, 100, l.BidSize, 1e-9)
	assert.InDelta(t, 162.3, l.BidPrice, 1e-9)
	assert.InDelta(t, 162.33, l.ReferencePrice, 1e-9)
	assert.InDelta(t, 162.36, l.AskPrice, 1e-9, "인덱스 6 이 가격이다")
	assert.InDelta(t, 100, l.AskSize, 1e-9, "인덱스 7 이 수량이다")
}

// 어느 길이에도 맞지 않으면 에러다.
func TestDecodeEquity_UnknownLength(t *testing.T) {
	_, err := decodeEquity(json.RawMessage(`["2026-09-06T13:30:00+00:00","AAPL",1,2]`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tiingo:")
}

func TestDecodeBOATS_QuoteAndTrade(t *testing.T) {
	// 문서 예시 그대로. mid = (81.58+81.59)/2 = 81.585 로 매핑이 검산된다.
	q := json.RawMessage(`["Q","2019-01-30T13:33:45.383129126-05:00",1548873225383129126,"vym",100,81.58,81.585,81.59,100]`)
	m, err := decodeBOATS(q)
	require.NoError(t, err)
	bq, ok := m.(BOATSQuote)
	require.True(t, ok)
	assert.Equal(t, "vym", bq.Ticker)
	assert.Equal(t, int64(1548873225383129126), bq.Nanoseconds)
	require.NotNil(t, bq.MidPrice)
	assert.InDelta(t, (bq.BidPrice+bq.AskPrice)/2, *bq.MidPrice, 1e-9)
	assert.InDelta(t, 81.59, bq.AskPrice, 1e-9, "인덱스 7 이 가격이다")
	assert.InDelta(t, 100, bq.AskSize, 1e-9, "인덱스 8 이 수량이다")

	// 문서 예시 그대로.
	tr := json.RawMessage(`["T","2019-01-30T13:33:45.594808294-05:00",1548873225594808294,"wes",50.285,200,"@","F","","X"]`)
	m, err = decodeBOATS(tr)
	require.NoError(t, err)
	b, ok := m.(BOATSTrade)
	require.True(t, ok)
	assert.False(t, b.Break, "T 는 정상 체결")
	assert.InDelta(t, 50.285, b.LastPrice, 1e-9)
	assert.InDelta(t, 200, b.LastSize, 1e-9)
	assert.Equal(t, [4]string{"@", "F", "", "X"}, b.SaleConditions)

	br := json.RawMessage(`["B","2019-01-30T13:33:45.594808294-05:00",1548873225594808294,"wes",50.285,200,"@","","",""]`)
	m, err = decodeBOATS(br)
	require.NoError(t, err)
	b2, ok := m.(BOATSTrade)
	require.True(t, ok)
	assert.True(t, b2.Break, "B 는 체결 취소")
}

// 한쪽 호가가 없으면 mid 는 null 이다 — 문서 명시.
func TestDecodeBOATS_NullMid(t *testing.T) {
	q := json.RawMessage(`["Q","2019-01-30T13:33:45+00:00",1548873225383129126,"vym",100,81.58,null,0,0]`)
	m, err := decodeBOATS(q)
	require.NoError(t, err)
	bq := m.(BOATSQuote)
	assert.Nil(t, bq.MidPrice)
}

// 모르는 종류 문자는 에러가 아니라 nil 이다 — 무시하고 다음 메시지로 간다.
func TestDecode_UnknownKindIsIgnored(t *testing.T) {
	m, err := decodeCrypto(json.RawMessage(`["Z","btcusd","2026-09-06T03:55:54+00:00","x",1,2]`))
	require.NoError(t, err)
	assert.Nil(t, m)

	m, err = decodeBOATS(json.RawMessage(`["Z","2019-01-30T13:33:45+00:00",1,"x"]`))
	require.NoError(t, err)
	assert.Nil(t, m)

	m, err = decodeForex(json.RawMessage(`["Z","eurusd","2026-09-06T03:55:54+00:00",1,2,3,4,5]`))
	require.NoError(t, err)
	assert.Nil(t, m)
}

// 짧은 배열은 에러로 올라온다.
func TestDecode_ShortArrayErrors(t *testing.T) {
	_, err := decodeCrypto(json.RawMessage(`["T","btcusd"]`))
	assert.Error(t, err)
	_, err = decodeForex(json.RawMessage(`["Q","eurusd"]`))
	assert.Error(t, err)
	_, err = decodeIEX(json.RawMessage(`["2026-09-06T13:30:00+00:00"]`))
	assert.Error(t, err)
	_, err = decodeBOATS(json.RawMessage(`["Q","2019-01-30T13:33:45+00:00",1,"vym"]`))
	assert.Error(t, err)
	_, err = decodeBOATS(json.RawMessage(`["T","2019-01-30T13:33:45+00:00",1,"wes"]`))
	assert.Error(t, err)
}

// 문서 예시 fixture 4종이 전부 디코딩돼야 한다.
func TestDecode_DocFixtures(t *testing.T) {
	cases := []struct {
		file   string
		decode func(json.RawMessage) (Message, error)
		want   int
	}{
		{"testdata/forex_doc.jsonl", decodeForex, 1},
		{"testdata/iex_doc.jsonl", decodeIEX, 1},
		{"testdata/equity_doc.jsonl", decodeEquity, 2},
		{"testdata/boats_doc.jsonl", decodeBOATS, 2},
	}
	for _, c := range cases {
		t.Run(c.file, func(t *testing.T) {
			raw, err := os.ReadFile(c.file)
			require.NoError(t, err)
			var n int
			for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
				if strings.TrimSpace(line) == "" {
					continue
				}
				e, err := parseEnvelope([]byte(line))
				require.NoError(t, err)
				m, err := c.decode(e.Data)
				require.NoError(t, err, line)
				require.NotNil(t, m, line)
				n++
			}
			assert.Equal(t, c.want, n)
		})
	}
}

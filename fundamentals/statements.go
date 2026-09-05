package fundamentals

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/kenshin579/tiingo-go/types"
)

// DataPoint 는 지표 하나의 코드와 값이다.
type DataPoint struct {
	DataCode string  `json:"dataCode"` // 지표 식별자(예: revenue). codes.go 의 상수 참고
	Value    float64 `json:"value"`    // 지표 값
}

// StatementData 는 한 기간의 재무 데이터를 네 묶음으로 나눈 것이다.
// 지표 집합은 기간마다 다르고 Tiingo 가 계속 추가하므로 코드→값 조회로 접근한다.
type StatementData struct {
	BalanceSheet    []DataPoint `json:"balanceSheet"`    // 재무상태표 지표
	IncomeStatement []DataPoint `json:"incomeStatement"` // 손익계산서 지표
	CashFlow        []DataPoint `json:"cashFlow"`        // 현금흐름표 지표
	Overview        []DataPoint `json:"overview"`        // 여러 재무제표를 조합한 비율·지표
}

// groups 는 네 묶음을 순회 순서대로 돌려준다.
func (s StatementData) groups() [][]DataPoint {
	return [][]DataPoint{s.IncomeStatement, s.BalanceSheet, s.CashFlow, s.Overview}
}

// Get 은 네 묶음 전체에서 코드를 찾는다. 없으면 ok 가 false 다(값 0 과 구분된다).
// 어느 묶음에 있는지 몰라도 되도록 전체를 훑는다 — 묶음을 특정하려면 필드를 직접 순회한다.
func (s StatementData) Get(code string) (float64, bool) {
	for _, g := range s.groups() {
		for _, dp := range g {
			if dp.DataCode == code {
				return dp.Value, true
			}
		}
	}
	return 0, false
}

// Map 은 코드→값 맵을 만든다. 여러 지표를 반복 조회할 때 쓴다.
func (s StatementData) Map() map[string]float64 {
	m := map[string]float64{}
	for _, g := range s.groups() {
		for _, dp := range g {
			m[dp.DataCode] = dp.Value
		}
	}
	return m
}

// Statement 는 한 회계기간의 재무제표다.
type Statement struct {
	Date          types.Date    `json:"date"`    // asReported=false 면 회계기간 종료일, true 면 SEC 공개일
	Year          int           `json:"year"`    // 회계연도
	Quarter       int           `json:"quarter"` // 0 이면 연간 보고서, 1~4 는 해당 분기
	StatementData StatementData `json:"statementData"`
}

// StatementOptions 는 Statements 의 선택 파라미터다. zero 값 필드는 요청에서 생략된다.
type StatementOptions struct {
	StartDate time.Time // 조회 시작일(이상, >=). zero 면 미전송
	EndDate   time.Time // 조회 종료일(이하, <=). zero 면 미전송
	Sort      string    // 정렬 컬럼. Statements 는 "date" / "-date" 만 지원한다
	// AsReported 가 true 면 SEC 공개 시점의 원본 수치를 받는다(date 는 공개일).
	// false(기본)면 수정·재작성을 반영한 최신 수치이고 date 는 회계기간 종료일이다.
	// 같은 기간을 조회해도 두 모드의 건수와 날짜가 다르다.
	AsReported bool
}

// params 는 옵션을 쿼리 맵으로 바꾼다. zero 값은 넣지 않는다.
func (o *StatementOptions) params() map[string]string {
	if o == nil {
		return nil
	}
	q := map[string]string{}
	if !o.StartDate.IsZero() {
		q["startDate"] = o.StartDate.Format(types.DateLayout)
	}
	if !o.EndDate.IsZero() {
		q["endDate"] = o.EndDate.Format(types.DateLayout)
	}
	if o.Sort != "" {
		q["sort"] = o.Sort
	}
	if o.AsReported {
		q["asReported"] = "true" // 기본값(false)일 때는 보내지 않는다
	}
	if len(q) == 0 {
		return nil
	}
	return q
}

// Statements 는 분기·연간 재무제표를 받는다. GET /tiingo/fundamentals/<ticker>/statements
// ticker 자리에는 티커 또는 permaTicker(상장폐지·재사용 심볼용)를 넣는다.
func (c *Client) Statements(ctx context.Context, ticker string, opts *StatementOptions) ([]Statement, error) {
	t := strings.TrimSpace(ticker)
	if t == "" {
		return nil, fmt.Errorf("tiingo: ticker must not be empty")
	}
	var ss []Statement
	path := "/tiingo/fundamentals/" + url.PathEscape(t) + "/statements"
	if err := c.http.GetJSON(ctx, path, opts.params(), &ss); err != nil {
		return nil, err
	}
	return ss, nil
}

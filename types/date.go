// Package types 는 tiingo-go 의 공용 값 타입이다.
// 루트 패키지와 카테고리 패키지(eod 등)가 함께 쓰기 때문에 import 순환을 피하려고 분리했다.
package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// DateLayout 은 Tiingo 요청 쿼리와 Date 직렬화에 쓰는 날짜 형식.
const DateLayout = "2006-01-02"

// Date 는 Tiingo 의 날짜 값이다. 응답에서 두 형식(RFC3339 타임스탬프, YYYY-MM-DD)이 모두
// 오기 때문에 둘 다 받아 time.Time 으로 정규화하고, 직렬화할 때는 YYYY-MM-DD 로 쓴다.
// null 이나 빈 문자열은 zero value 가 되며 IsZero() 로 구분한다.
// 날짜 전용 타입이다 — 갱신 시각(statementLastUpdated)이나 체결 시각처럼 시각까지 의미 있는
// 필드는 Time 을 쓴다.
type Date struct {
	time.Time
}

// parse 는 RFC3339 타임스탬프와 YYYY-MM-DD 를 모두 받는다.
// 빈 문자열은 zero value 다. UnmarshalJSON 과 UnmarshalText 가 함께 쓴다.
//
// 오프셋이 있는 타임스탬프는 UTC 로 환산하지 않고 **표기된 벽시계 날짜**를 취한다.
// 날짜 전용 타입이므로 "2019-01-02T19:00:00-05:00" 은 2019-01-02 여야지, UTC 환산 결과인
// 2019-01-03 이 되면 안 된다(하루 밀림).
func (d *Date) parse(s string) error {
	if s == "" {
		d.Time = time.Time{}
		return nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		d.Time = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		return nil
	}
	t, err := time.Parse(DateLayout, s)
	if err != nil {
		return fmt.Errorf("tiingo: unrecognized date %q", s)
	}
	d.Time = t
	return nil
}

// UnmarshalJSON 은 RFC3339 타임스탬프와 YYYY-MM-DD 를 모두 받는다.
func (d *Date) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		d.Time = time.Time{}
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("tiingo: date must be a JSON string: %w", err)
	}
	return d.parse(s)
}

// MarshalJSON 은 YYYY-MM-DD 문자열로 쓴다. zero value 는 null.
func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.Time.Format(DateLayout))
}

// MarshalText 는 MarshalJSON 과 같은 YYYY-MM-DD 규칙을 따른다(map 키·yaml 등 텍스트 인코더용).
// 이 메서드가 없으면 임베드된 time.Time 의 것이 승격돼 RFC3339 로 나간다.
func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalText 는 UnmarshalJSON 과 같은 두 형식을 받는다.
func (d *Date) UnmarshalText(b []byte) error {
	return d.parse(string(b))
}

// String 은 YYYY-MM-DD 를 반환한다. zero value 는 빈 문자열.
func (d Date) String() string {
	if d.Time.IsZero() {
		return ""
	}
	return d.Time.Format(DateLayout)
}

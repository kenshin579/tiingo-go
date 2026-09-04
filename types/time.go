package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// Time 은 시각까지 의미 있는 Tiingo 타임스탬프다(예: statementLastUpdated).
// 날짜만 필요한 필드는 Date 를 쓴다 — Date 는 시각을 버리고 날짜만 남긴다.
// null 이나 빈 문자열은 zero value 가 되며 IsZero() 로 구분한다.
type Time struct {
	time.Time
}

// parse 는 RFC3339 를 UTC 로 파싱한다. 빈 문자열은 zero value 다.
func (t *Time) parse(s string) error {
	if s == "" {
		t.Time = time.Time{}
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return fmt.Errorf("tiingo: unrecognized timestamp %q", s)
	}
	t.Time = parsed.UTC()
	return nil
}

// UnmarshalJSON 은 RFC3339 문자열을 받는다.
func (t *Time) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		t.Time = time.Time{}
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("tiingo: timestamp must be a JSON string: %w", err)
	}
	return t.parse(s)
}

// UnmarshalText 는 UnmarshalJSON 과 같은 형식을 받는다.
func (t *Time) UnmarshalText(b []byte) error { return t.parse(string(b)) }

// MarshalJSON 은 RFC3339 문자열로 쓴다. zero value 는 null.
func (t Time) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.String())
}

// MarshalText 는 RFC3339 문자열로 쓴다. zero value 는 빈 문자열.
func (t Time) MarshalText() ([]byte, error) { return []byte(t.String()), nil }

// String 은 RFC3339 를 반환한다. zero value 는 빈 문자열.
func (t Time) String() string {
	if t.Time.IsZero() {
		return ""
	}
	return t.Time.Format(time.RFC3339Nano)
}

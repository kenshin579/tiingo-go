package stream

import (
	"encoding/json"
	"fmt"

	"github.com/kenshin579/tiingo-go/types"
)

// arr 은 위치 기반 data 배열을 읽는 커서다.
//
// Tiingo WebSocket 은 데이터를 필드명이 아니라 배열 위치로 보낸다. 위치 매핑이 틀리면
// 컴파일도 테스트도 통과하고 값만 조용히 바뀌므로, 길이와 타입을 여기서 확인해 시끄럽게
// 실패시키는 것이 유일한 방어선이다. 첫 에러를 붙들고 있다가 err() 로 한 번에 돌려준다.
type arr struct {
	raw   []json.RawMessage
	first error // 처음 발생한 에러. 이후 실패는 덮어쓰지 않는다
}

// newArr 은 배열 원문을 커서로 만든다. want 보다 짧으면 에러다.
//
// want 보다 긴 것은 허용한다 — Tiingo 가 REST 에서 "필드를 계속 추가하겠다" 고 밝혔고
// WebSocket 도 같을 수 있다. 뒤에 붙는 원소는 무시한다.
func newArr(raw json.RawMessage, want int) (*arr, error) {
	var xs []json.RawMessage
	if err := json.Unmarshal(raw, &xs); err != nil {
		return nil, fmt.Errorf("tiingo: data must be a JSON array: %w", err)
	}
	if len(xs) < want {
		return nil, fmt.Errorf("tiingo: data array too short: want at least %d elements, got %d", want, len(xs))
	}
	return &arr{raw: xs}, nil
}

// arrFrom 은 이미 파싱된 원소 목록으로 커서를 만든다. 길이 검사는 호출자가 이미 했을 때 쓴다.
func arrFrom(xs []json.RawMessage) *arr { return &arr{raw: xs} }

// fail 은 첫 에러만 기록한다.
func (a *arr) fail(i int, err error) {
	if a.first == nil {
		a.first = fmt.Errorf("tiingo: data index %d: %w", i, err)
	}
}

// err 은 지금까지 발생한 첫 에러를 돌려준다.
func (a *arr) err() error { return a.first }

// str 은 문자열 원소를 읽는다.
func (a *arr) str(i int) string {
	if string(a.raw[i]) == "null" {
		a.fail(i, fmt.Errorf("unexpected null"))
		return ""
	}
	var v string
	if err := json.Unmarshal(a.raw[i], &v); err != nil {
		a.fail(i, err)
		return ""
	}
	return v
}

// f64 는 실수 원소를 읽는다. null 은 에러다 — 결손을 허용하려면 f64p 를 쓴다.
//
// json.Unmarshal 은 null 을 float64 에 조용히 무시하므로(0 유지, 에러 없음) null 검사를
// 직접 해야 한다.
func (a *arr) f64(i int) float64 {
	if string(a.raw[i]) == "null" {
		a.fail(i, fmt.Errorf("unexpected null"))
		return 0
	}
	var v float64
	if err := json.Unmarshal(a.raw[i], &v); err != nil {
		a.fail(i, err)
		return 0
	}
	return v
}

// f64p 는 실수 원소를 읽되 null 을 nil 로 돌려준다.
func (a *arr) f64p(i int) *float64 {
	if string(a.raw[i]) == "null" {
		return nil
	}
	v := a.f64(i)
	if a.first != nil {
		return nil
	}
	return &v
}

// i64 는 정수 원소를 읽는다.
func (a *arr) i64(i int) int64 {
	if string(a.raw[i]) == "null" {
		a.fail(i, fmt.Errorf("unexpected null"))
		return 0
	}
	var v int64
	if err := json.Unmarshal(a.raw[i], &v); err != nil {
		a.fail(i, err)
		return 0
	}
	return v
}

// time 은 타임스탬프 원소를 읽는다.
func (a *arr) time(i int) types.Time {
	if string(a.raw[i]) == "null" {
		a.fail(i, fmt.Errorf("unexpected null"))
		return types.Time{}
	}
	var v types.Time
	if err := json.Unmarshal(a.raw[i], &v); err != nil {
		a.fail(i, err)
		return types.Time{}
	}
	return v
}

package search

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FIGI 는 OpenFIGI 식별자다.
//
// Tiingo 는 값이 없을 때 JSON null 과 문자열 "nan" 두 가지를 섞어 보낸다(둘 다 실제 응답에서 확인).
// "nan" 은 FIGI 형식이 아니라 결손을 뜻하는 데이터 오류이므로 null 과 함께 빈 문자열로 정규화한다 —
// 호출자는 `figi != ""` 하나만 확인하면 된다.
// 기저 타입이 string 이라 비교·출력은 문자열과 똑같이 쓴다.
// 식별자라 앞뒤 공백은 의미가 없어 잘라낸다 — 관측된 사례가 있어서가 아니라 값의 성격상 안전하다.
//
// MarshalJSON 은 두지 않았다. 그래서 Result 를 다시 직렬화하면 Tiingo 가 null 을 보냈던 자리에
// 빈 문자열이 나간다 — 결손을 한 가지로 모으는 게 목적이라 의도한 비대칭이다.
type FIGI string

// missingFIGI 는 Tiingo 가 결손을 표현하는 문자열이다. 대소문자는 구분하지 않는다.
const missingFIGI = "nan"

// UnmarshalJSON 은 null 과 "nan" 을 빈 문자열로 정규화한다.
// 문자열이 아닌 값(숫자·객체 등)은 에러다 — 응답 스키마가 바뀐 것이므로 조용히 넘기지 않는다.
func (f *FIGI) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*f = ""
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("tiingo: figi must be a JSON string: %w", err)
	}
	s = strings.TrimSpace(s)
	if strings.EqualFold(s, missingFIGI) {
		*f = ""
		return nil
	}
	*f = FIGI(s)
	return nil
}

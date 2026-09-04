// Package tiingo 는 Tiingo API 의 Go 클라이언트다.
package tiingo

import "github.com/kenshin579/tiingo-go/types"

// DateLayout 은 Tiingo 요청 쿼리와 Date 직렬화에 쓰는 날짜 형식.
const DateLayout = types.DateLayout

// Date 는 Tiingo 의 날짜 값이다. 응답에서 두 형식(RFC3339 타임스탬프, YYYY-MM-DD)이 모두
// 오기 때문에 둘 다 받아 time.Time 으로 정규화하고, 직렬화할 때는 YYYY-MM-DD 로 쓴다.
// null 이나 빈 문자열은 zero value 가 되며 IsZero() 로 구분한다.
// types.Date 의 별칭이라 카테고리 패키지(eod 등)와 같은 타입이다.
type Date = types.Date

// Time 은 시각까지 의미 있는 Tiingo 타임스탬프다(예: statementLastUpdated).
// 날짜만 필요한 필드는 Date 를 쓴다 — Date 는 시각을 버리고 날짜만 남긴다.
// null 이나 빈 문자열은 zero value 가 되며 IsZero() 로 구분한다.
// types.Time 의 별칭이라 카테고리 패키지(fundamentals 등)와 같은 타입이다.
type Time = types.Time

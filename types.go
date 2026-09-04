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

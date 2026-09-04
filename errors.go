package tiingo

import "github.com/kenshin579/tiingo-go/internal/httpclient"

// APIError 는 Tiingo 에러 응답이다. errors.As 로 StatusCode/Message 에 접근한다.
type APIError = httpclient.APIError

// ErrNotFound 는 조회 결과가 없을 때 서비스 계층이 반환한다.
var ErrNotFound = httpclient.ErrNotFound

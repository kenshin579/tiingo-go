package tiingo

import (
	"errors"
	"os"
)

// APIKeyEnv 는 API 키를 읽는 환경변수 이름.
const APIKeyEnv = "TIINGO_API_KEY"

// NewClientFromEnv 는 TIINGO_API_KEY 환경변수로 Client 를 만든다.
func NewClientFromEnv(opts ...Option) (*Client, error) {
	key := os.Getenv(APIKeyEnv)
	if key == "" {
		return nil, errors.New("tiingo: " + APIKeyEnv + " is not set")
	}
	return NewClient(key, opts...)
}

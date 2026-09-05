package crypto

import (
	"fmt"
	"strings"
)

// joinTickers 는 티커 목록을 검증해 콤마로 합친다.
// 원소가 공백뿐이면 에러다. required 가 true 면 목록 자체가 비어도 에러다.
func joinTickers(tickers []string, required bool) (string, error) {
	cleaned := make([]string, 0, len(tickers))
	for _, t := range tickers {
		t = strings.TrimSpace(t)
		if t == "" {
			return "", fmt.Errorf("tiingo: ticker must not be empty")
		}
		cleaned = append(cleaned, t)
	}
	if required && len(cleaned) == 0 {
		return "", fmt.Errorf("tiingo: at least one ticker is required")
	}
	return strings.Join(cleaned, ","), nil
}

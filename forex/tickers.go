package forex

import (
	"fmt"
	"net/url"
	"strings"
)

// cleanTickers 는 티커 목록을 다듬고 검증한다. 공백뿐인 원소나 빈 목록은 에러다.
func cleanTickers(tickers []string) ([]string, error) {
	cleaned := make([]string, 0, len(tickers))
	for _, t := range tickers {
		t = strings.TrimSpace(t)
		if t == "" {
			return nil, fmt.Errorf("tiingo: ticker must not be empty")
		}
		cleaned = append(cleaned, t)
	}
	if len(cleaned) == 0 {
		return nil, fmt.Errorf("tiingo: at least one ticker is required")
	}
	return cleaned, nil
}

// joinTickers 는 쿼리 파라미터용으로 콤마 결합한다(TopOfBook).
func joinTickers(tickers []string) (string, error) {
	cleaned, err := cleanTickers(tickers)
	if err != nil {
		return "", err
	}
	return strings.Join(cleaned, ","), nil
}

// pathTickers 는 URL 경로용으로 결합한다(Prices).
// 콤마는 경로 구분자로 살려야 하므로 티커를 개별 escape 한 뒤 콤마로 잇는다.
// (먼저 합쳐서 escape 하면 콤마가 %2C 가 된다 — API 는 받아주지만 URL 이 읽기 어렵다.)
func pathTickers(tickers []string) (string, error) {
	cleaned, err := cleanTickers(tickers)
	if err != nil {
		return "", err
	}
	escaped := make([]string, len(cleaned))
	for i, t := range cleaned {
		escaped[i] = url.PathEscape(t)
	}
	return strings.Join(escaped, ","), nil
}

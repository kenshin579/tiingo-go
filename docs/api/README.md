# Tiingo API 문서 카탈로그

Tiingo 문서 사이트(https://www.tiingo.com/documentation) 를 자동 변환한 md 총 23개 페이지(tools/gendocs)와
Tiingo 가 공식 제공하는 llms 원본 2개를 보관합니다. `tiingo-go` SDK 개발의 1차 참조입니다.

- 페이지 md: 웹 문서를 페이지 단위로 변환. 엔드포인트별 Request/Response 필드 표(타입·설명)와 Python 예시 + 응답 예시 JSON 포함.
- llms 원본: Tiingo 가 유지하는 개념·정책·플랜 제한·심볼 규칙의 source of truth. 일부 응답 필드는 타입 없이 이름만 나열돼 있어 페이지 md 로 보완.

## 공식 원본 (Tiingo 제공)

| 파일 | 출처 | Last updated | 가져온 날짜 |
| --- | --- | --- | --- |
| `llms.txt` | https://www.tiingo.com/llms.txt | 2026-08-18 | 2026-09-04 |
| `llms-full.txt` | https://www.tiingo.com/llms-full.txt | 2026-08-18 | 2026-09-04 |

## 재생성

```bash
./scripts/fetch-docs.sh                      # llms.txt / llms-full.txt 갱신 (curl)
cd tools/gendocs && npm install && npx playwright install chromium && npm run gen   # 페이지 md 재생성
```

## 페이지

### 1. General

- [1.1 Overview](general/overview.md)
- [1.2 Connecting](general/connecting.md)
- [1.3 Changelog](general/changelog.md)

### 2. REST

- [2.1 End-of-Day](rest/end-of-day.md)
- [2.2 News](rest/news.md)
- [2.3 Crypto](rest/crypto.md)
- [2.4 Forex](rest/forex.md)
- [2.5 Equity Realtime](rest/equity-realtime-stock-data.md)
- [2.6 IEX](rest/iex.md)
- [2.7 BOATS Overnight](rest/boats.md)
- [2.8 Fundamentals](rest/fundamentals.md)
- [2.9 Fund Fees](rest/mutual-fund-and-etf-fees.md)
- [2.10 Dividends](rest/dividends.md)
- [2.11 Splits](rest/splits.md)

### 3. Websockets

- [3.1 Crypto](websockets/crypto.md)
- [3.2 Forex](websockets/forex.md)
- [3.3 Equity Realtime](websockets/equity-realtime-stock-data.md)
- [3.4 IEX](websockets/iex.md)
- [3.5 BOATS Overnight](websockets/boats.md)

### 4. Utilities

- [4.1 Search](utilities/search.md)

### 5. Appendix

- [5.1 Developer Program](appendix/developers.md)
- [5.2 Integrations](appendix/integrations.md)
- [5.3 Symbology](appendix/symbology.md)

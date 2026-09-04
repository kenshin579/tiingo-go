# tiingo-go

[Tiingo](https://www.tiingo.com) 금융 데이터 API(주식 EOD/실시간 시세, 뉴스, 펀더멘털, 암호화폐, 외환)의 Go 클라이언트 라이브러리. **구현 예정** — 현재는 API 문서 카탈로그 단계입니다.

## 문서

- [`docs/api/README.md`](docs/api/README.md) — Tiingo 문서 사이트 23페이지를 변환한 md + Tiingo 공식 `llms.txt`/`llms-full.txt` 원본.
- 재생성: `./scripts/fetch-docs.sh`(llms 원본), `cd tools/gendocs && npm install && npx playwright install chromium && npm run gen`(페이지 md).
- 설계: [`docs/superpowers/specs/`](docs/superpowers/specs/)

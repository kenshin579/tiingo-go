#!/usr/bin/env bash
# fetch-docs.sh — Tiingo 공식 llms 원본 2개를 docs/api/ 로 가져온다.
#
#   ./scripts/fetch-docs.sh
#
# 동작:
#   1. https://www.tiingo.com/llms.txt, llms-full.txt 를 임시 디렉토리에 다운로드
#      — 하나라도 실패하면 아무것도 반영하지 않고 중단
#   2. 내용이 바뀐 파일만 docs/api/ 로 교체
#   3. 교체된 파일에 대해 docs/api/README.md 의 "공식 원본" 표 행(Last updated·가져온 날짜) 갱신
#      (Last updated 는 파일 안의 'Last updated: YYYY-MM-DD' 또는 'last_updated: "YYYY-MM-DD"' 첫 등장값, 없으면 '-')
#
# 의존: curl
# TIINGO_DOCS_BASE 환경변수로 원본 base URL 을 바꿀 수 있다(테스트용).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="$ROOT/docs/api"
README="$OUT/README.md"
BASE="${TIINGO_DOCS_BASE:-https://www.tiingo.com}"
UA="Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36"
FILES=(llms.txt llms-full.txt)

if ! command -v curl >/dev/null 2>&1; then
  echo "error: 'curl' 가 필요합니다" >&2
  exit 1
fi

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

# 1) 전부 다운로드 (하나라도 실패하면 set -e 로 여기서 중단 → docs/api 미변경)
for f in "${FILES[@]}"; do
  echo "GET $BASE/$f"
  curl -fsSL -A "$UA" "$BASE/$f" -o "$TMP/$f"
  if [[ ! -s "$TMP/$f" ]]; then
    echo "error: $f 가 비어 있음" >&2
    exit 1
  fi
done

TODAY="$(date +%Y-%m-%d)"
mkdir -p "$OUT"

# last_updated <file> — 'Last updated: YYYY-MM-DD'(llms.txt) 또는 'last_updated: "YYYY-MM-DD"'(llms-full.txt
# 프론트매터) 첫 등장값의 날짜, 없으면 빈 문자열(호출부에서 '-')
last_updated() {
  grep -m1 -oiE 'last[ _]updated:? *"?[0-9]{4}-[0-9]{2}-[0-9]{2}' "$1" | grep -oE '[0-9]{4}-[0-9]{2}-[0-9]{2}' || true
}

# update_readme_row <name> <updated> — README 원본 표에서 해당 행의 Last updated·가져온 날짜를 바꾼다
update_readme_row() {
  local name="$1" updated="$2"
  if [[ ! -f "$README" ]]; then
    echo "  warn: $README 가 없어 원본 표를 갱신하지 못함 (tools/gendocs 로 먼저 생성)" >&2
    return 0
  fi
  local esc="${name//./\\.}"
  sed -E "s#^(\| \`$esc\` \| [^|]* \|) [^|]* \| [^|]* \|#\1 $updated | $TODAY |#" "$README" > "$TMP/README.md"
  mv "$TMP/README.md" "$README"
  if ! grep -qF "| \`$name\` | $BASE/$name | $updated | $TODAY |" "$README"; then
    echo "  warn: README 원본 표에서 $name 행을 찾지 못해 갱신되지 않음" >&2
  fi
}

# 2), 3)
for f in "${FILES[@]}"; do
  if [[ -f "$OUT/$f" ]] && cmp -s "$TMP/$f" "$OUT/$f"; then
    echo "  $f: 변경 없음"
    continue
  fi
  mv "$TMP/$f" "$OUT/$f"
  upd="$(last_updated "$OUT/$f")"
  [[ -n "$upd" ]] || upd="-"
  update_readme_row "$f" "$upd"
  echo "  $f: 갱신 (Last updated=$upd)"
done

echo "done → $OUT"

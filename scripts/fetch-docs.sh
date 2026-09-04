#!/usr/bin/env bash
# fetch-docs.sh — Tiingo 공식 llms 원본 2개를 docs/api/ 로 가져온다.
#
#   ./scripts/fetch-docs.sh
#
# 동작:
#   1. https://www.tiingo.com/llms.txt, llms-full.txt 를 임시 디렉토리에 다운로드
#      — 하나라도 실패하면 아무것도 반영하지 않고 중단
#   2. 다운로드된(임시) 파일에서 Last updated 날짜를 전부 추출·검증 — mv 전에 수행하여
#      하나라도 이상하면 아무것도 반영하지 않고 중단
#   3. 내용이 바뀐 파일만 docs/api/ 로 교체
#   4. 교체된 파일에 대해 docs/api/README.md 의 "공식 원본" 표 행(Last updated·가져온 날짜) 갱신
#      (Last updated 는 파일 안의 'Last updated: YYYY-MM-DD' 또는 'last_updated: "YYYY-MM-DD"' 첫 등장값, 없으면 '-')
#
# 의존: curl
# TIINGO_DOCS_BASE 환경변수로 원본 base URL 을 바꿀 수 있다(테스트용).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT="$ROOT/docs/api"
README="$OUT/README.md"
BASE="${TIINGO_DOCS_BASE:-https://www.tiingo.com}"
# 브라우저 UA — 일부 CDN/WAF 가 curl 기본 UA 를 차단하는 경우 대비(현재 Tiingo 는 기본 UA 도 200)
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
  curl -fsSL --max-time 60 --retry 2 -A "$UA" "$BASE/$f" -o "$TMP/$f"
  if [[ ! -s "$TMP/$f" ]]; then
    echo "error: $f 가 비어 있음" >&2
    exit 1
  fi
done

TODAY="$(date +%Y-%m-%d)"
mkdir -p "$OUT"

# last_updated <file> — 'Last updated: YYYY-MM-DD'(llms.txt) 또는 'last_updated: "YYYY-MM-DD"'(llms-full.txt
# 프론트매터) 첫 등장값의 날짜, 없으면 빈 문자열(호출부에서 '-')
# 주의: -m1 은 "매치 라인 수"를 제한할 뿐, 한 줄에 날짜가 2개 이상 있으면(예: "Last updated:
# 2026-09-01 (prev last_updated: 2026-08-01)") 두 번째 grep -oE 가 둘 다 출력할 수 있으므로
# 그쪽도 -m1 로 첫 번째 값만 취한다.
last_updated() {
  grep -m1 -oiE 'last[ _]updated:? *"?[0-9]{4}-[0-9]{2}-[0-9]{2}' "$1" | grep -m1 -oE '[0-9]{4}-[0-9]{2}-[0-9]{2}' || true
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
  # URL(출처) 컬럼은 비교하지 않는다 — TIINGO_DOCS_BASE 로 테스트용 base 를 쓰면 README 의
  # 출처 컬럼(실제 공식 URL)과 달라져 허위 warn 이 뜨기 때문
  if ! grep -qE "^\| \`$esc\` \| [^|]* \| $updated \| $TODAY \|" "$README"; then
    echo "  warn: README 원본 표에서 $name 행을 찾지 못해 갱신되지 않음" >&2
  fi
}

# 2) 날짜 추출·검증 (mv 전에 전부 — 하나라도 이상하면 아무것도 반영하지 않음)
declare -a UPDATED=()
for f in "${FILES[@]}"; do
  upd="$(last_updated "$TMP/$f")"
  [[ -n "$upd" ]] || upd="-"
  if [[ "$upd" != "-" && ! "$upd" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]]; then
    echo "error: $f 의 Last updated 값이 예상 밖: '$upd'" >&2
    exit 1
  fi
  UPDATED+=("$upd")
done

# 3), 4) 내용이 바뀐 파일만 교체 + README 행 갱신
for i in "${!FILES[@]}"; do
  f="${FILES[$i]}"; upd="${UPDATED[$i]}"
  if [[ -f "$OUT/$f" ]] && cmp -s "$TMP/$f" "$OUT/$f"; then
    echo "  $f: 변경 없음"
    continue
  fi
  mv "$TMP/$f" "$OUT/$f"
  update_readme_row "$f" "$upd"
  echo "  $f: 갱신 (Last updated=$upd)"
done

echo "done → $OUT"

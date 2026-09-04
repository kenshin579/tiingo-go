package fundamentals

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// fixture 의 dataCode 가 전부 상수로 존재하는지 확인한다.
// Tiingo 가 지표를 추가하면 fixture 갱신 시 이 테스트가 상수 누락을 잡는다.
func TestCodes_CoverFixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/definitions.json")
	require.NoError(t, err)
	var ds []Definition
	require.NoError(t, json.Unmarshal(raw, &ds))
	require.NotEmpty(t, ds)

	known := map[string]bool{}
	for _, c := range AllCodes {
		known[c] = true
	}
	var missing []string
	for _, d := range ds {
		if !known[d.DataCode] {
			missing = append(missing, d.DataCode)
		}
	}
	require.Emptyf(t, missing, "codes.go 에 없는 dataCode: %v", missing)
	require.Len(t, AllCodes, len(ds), "상수 개수와 정의 개수가 같아야 한다")
}

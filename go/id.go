// Package niface は niface 規格(specVersion 1)の型と item id 導出の参照実装。
//
// id 導出: id = lowercase-hex( sha256( JCS( identity ) ) )
// JCS は RFC 8785 のサブセット実装(文字列/整数/bool/null/配列/オブジェクト)。
// 浮動小数は ECMAScript 数値表記が必要なため本サブセットでは非対応とし、
// エラーを返す(規格上 key に非決定的な値を含めないため実用上は十分)。
package niface

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"
)

// Identity は item id の導出元。
type Identity struct {
	Kind string `json:"kind"`
	Key  any    `json:"key"`
}

// DeriveID は identity から item id を導出する。
func DeriveID(id Identity) (string, error) {
	m := map[string]any{"kind": id.Kind, "key": id.Key}
	c, err := canonicalize(m)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(c))
	return hex.EncodeToString(sum[:]), nil
}

// canonicalize は RFC 8785 (JCS) サブセットの正準化を行う。
func canonicalize(v any) (string, error) {
	switch x := v.(type) {
	case nil:
		return "null", nil
	case bool:
		if x {
			return "true", nil
		}
		return "false", nil
	case string:
		return encodeString(x), nil
	case int:
		return strconv.FormatInt(int64(x), 10), nil
	case int64:
		return strconv.FormatInt(x, 10), nil
	case float64:
		// 整数値の float は整数として表記(JCS 準拠)。非整数はサブセット外。
		if x == math.Trunc(x) && !math.IsInf(x, 0) && math.Abs(x) < 1e15 {
			return strconv.FormatInt(int64(x), 10), nil
		}
		return "", fmt.Errorf("niface: non-integral number %v is unsupported in JCS subset", x)
	case []any:
		parts := make([]string, len(x))
		for i, e := range x {
			c, err := canonicalize(e)
			if err != nil {
				return "", err
			}
			parts[i] = c
		}
		return "[" + strings.Join(parts, ",") + "]", nil
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		// JCS のキー順序は UTF-16 code unit 順
		sort.Slice(keys, func(i, j int) bool {
			return lessUTF16(keys[i], keys[j])
		})
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			c, err := canonicalize(x[k])
			if err != nil {
				return "", err
			}
			parts = append(parts, encodeString(k)+":"+c)
		}
		return "{" + strings.Join(parts, ",") + "}", nil
	default:
		return "", fmt.Errorf("niface: unsupported type %T in JCS subset", v)
	}
}

func lessUTF16(a, b string) bool {
	ua, ub := utf16.Encode([]rune(a)), utf16.Encode([]rune(b))
	for i := 0; i < len(ua) && i < len(ub); i++ {
		if ua[i] != ub[i] {
			return ua[i] < ub[i]
		}
	}
	return len(ua) < len(ub)
}

// encodeString は JCS の最小エスケープで文字列をエンコードする。
// 非 ASCII は UTF-8 のまま出力する(\u エスケープしない)。
func encodeString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				fmt.Fprintf(&b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}

// Package niface は niface 規格(specVersion 1)の型と item id 導出の参照実装。
//
// id 導出: id = lowercase-hex( sha256( JCS( identity ) ) )
// JCS は RFC 8785。identity の値域は spec §5 に定める制約に従う:
// 文字列(全 Unicode)/ 整数(±(2^53−1))/ bool / null / 配列 /
// オブジェクト(メンバー名は ASCII)。非整数の数値・範囲外の整数・非 ASCII の
// メンバー名は域外として拒否する(エラーを返す)。制約された値域の上では
// 本実装はフル JCS と一致する。
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

// maxSafeInteger は identity 整数値の値域上限 2^53−1(spec §5)。
// IEEE 754 倍精度が正確に表せる整数の上限に一致させ、int / int64 / float64 の
// 各経路に同一の範囲ガードを課す。
const maxSafeInteger = 1<<53 - 1

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
		if int64(x) > maxSafeInteger || int64(x) < -maxSafeInteger {
			return "", fmt.Errorf("niface: integer %d is out of identity domain ±(2^53-1)", x)
		}
		return strconv.FormatInt(int64(x), 10), nil
	case int64:
		if x > maxSafeInteger || x < -maxSafeInteger {
			return "", fmt.Errorf("niface: integer %d is out of identity domain ±(2^53-1)", x)
		}
		return strconv.FormatInt(x, 10), nil
	case float64:
		// 整数値かつ安全整数域内の float は整数として表記(JCS 準拠)。
		// 非整数・範囲外・NaN/Inf は域外(spec §5)。
		if x == math.Trunc(x) && !math.IsInf(x, 0) && math.Abs(x) <= maxSafeInteger {
			return strconv.FormatInt(int64(x), 10), nil
		}
		return "", fmt.Errorf("niface: number %v is out of identity domain (non-integral or |x| > 2^53-1)", x)
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
			if !isASCII(k) {
				return "", fmt.Errorf("niface: object member name %q must be ASCII in identity domain", k)
			}
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

// isASCII は s の全バイトが ASCII(U+0000–U+007F)かを返す。
// identity のオブジェクトメンバー名の値域検査に用いる(spec §5)。
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
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

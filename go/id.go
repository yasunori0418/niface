// Package niface は niface 規格(specVersion 1)の型と item id 導出の参照実装。
//
// id 導出: id = lowercase-hex( sha256( JCS( identity ) ) )
// JCS は RFC 8785。identity の値域は spec §5 に定める制約に従う:
// 文字列(全 Unicode)/ 整数(±(2^53−1))/ bool / null / 配列 /
// オブジェクト(メンバー名は ASCII)。数値は表記で判定し、浮動小数点(float64)・
// 小数点/指数表記(1.0・1e3 等、値が整数でも)・範囲外の整数・非 ASCII の
// メンバー名は域外として拒否する(エラーを返す)。制約された値域の上では
// 本実装はフル JCS と一致する。
package niface

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"
)

// maxSafeInteger は identity 整数値の値域上限 2^53−1(spec §5)。
// IEEE 754 倍精度が正確に表せる整数の上限に一致させ、int / int64 / json.Number の
// 各経路に同一の範囲ガードを課す。
const maxSafeInteger = 1<<53 - 1

// 値域違反を分類する sentinel error(errors.Is で判定可能)。具体値は %w で wrap する。
var (
	errFloatType    = errors.New("niface: floating-point number is not allowed in identity domain")
	errNonInteger   = errors.New("niface: number must use integer notation (no fraction/exponent)")
	errIntegerRange = errors.New("niface: integer is out of identity domain ±(2^53-1)")
	errNonASCIIKey  = errors.New("niface: object member name must be ASCII")
	errUnsupported  = errors.New("niface: unsupported type in identity domain")
)

// checkIntRange は整数値 n が identity の値域 ±(2^53−1)(spec §5)に収まるかを
// 検査する。域外なら errIntegerRange を wrap して返す。int / int64 / json.Number の
// 各経路が共通で用い、境界判定を一箇所に集約する。
func checkIntRange(n int64) error {
	if n > maxSafeInteger || n < -maxSafeInteger {
		return fmt.Errorf("%w (got integer %d)", errIntegerRange, n)
	}
	return nil
}

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

// canonicalize は identity 値を spec §5 の値域制約のもとで JCS (RFC 8785) 正準化する。
// 域外(浮動小数・非整数表記・範囲外整数・非 ASCII メンバー名・未対応型)はエラーを返す。
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
		if err := checkIntRange(int64(x)); err != nil {
			return "", err
		}
		return strconv.FormatInt(int64(x), 10), nil
	case int64:
		if err := checkIntRange(x); err != nil {
			return "", err
		}
		return strconv.FormatInt(x, 10), nil
	case json.Number:
		// JSON 数値は表記で判定する(spec §5)。小数点・指数を持つ表記
		// (1.0・1e3 等、値が整数でも)は域外。整数表記のみ ±(2^53−1) で受理。
		s := x.String()
		if strings.ContainsAny(s, ".eE") {
			return "", fmt.Errorf("%w (got %q)", errNonInteger, s)
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			// int64 に収まらない整数表記(例: 99999999999999999999)。
			// 値域外なので errIntegerRange に分類する。
			return "", fmt.Errorf("%w (got integer %q)", errIntegerRange, s)
		}
		if err := checkIntRange(n); err != nil {
			return "", err
		}
		return strconv.FormatInt(n, 10), nil
	case float64:
		// 浮動小数点数は identity に使えない(整数は int/int64/json.Number で表す)。
		// encoding/json は 1 と 1.0 を共に float64 にするため、float64 型は表記情報を
		// 持たず整数か判定できない。よって型として拒否する(spec §5・表記で判定)。
		return "", fmt.Errorf("%w (got float64 %v)", errFloatType, x)
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
				return "", fmt.Errorf("%w (got %q)", errNonASCIIKey, k)
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
		return "", fmt.Errorf("%w (got %T)", errUnsupported, v)
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

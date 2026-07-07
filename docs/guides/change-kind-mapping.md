# 動作系ツールの change.kind マッピング指針

nherd のようにサービス・プロセスの動作（start / stop / restart / reload）を扱う「動作系」ツールが、動作を `change.kind`（`add` / `remove` / `modify`）へどうマップするかの指針。本書は規範ではなく指針であり、規範は `spec/v1/spec.md`（§3, §4）にある。

## 前提

`change.kind` は配置系の差分語彙（`add` / `remove` / `modify`）であり、動作系のために enum へ新値を追加しない。未知の enum 値は must-ignore で救えず消費側の kind 分岐を壊すため、追加は実質非互換になる（status 新値を棄却したのと同じ理由 → ADR-0020）。動作系ツールは既存 3 値へマップし、3 値で表しきれない意味は `info` で運ぶ（→ ADR-0007）。

## マッピング

| 動作 | kind | info の例 |
|------|------|-----------|
| start（停止 → 稼働） | `add` | — |
| stop（稼働 → 停止） | `remove` | — |
| restart | `modify` | `{ "action": "restart" }` |
| reload | `modify` | `{ "action": "reload" }` |

- start / stop は「稼働状態」の追加 / 除去として `add` / `remove` にマップする
- restart / reload は稼働という状態を保ったままの遷移なので `modify` にマップし、どの動作かは `info.action` で運ぶ。kind しか見ない消費側には「変更があった」ことが、動作系を知る消費側には `info.action` で具体が伝わる

## reversible の判断

`reversible` は「この差分単位で巻き戻し可能か」（→ `spec/v1/spec.md` §4）。動作系では次のように判断する。

- restart / reload は原則 `false`。巻き戻しに相当する操作（再度の restart）は再起動前の runtime 状態（メモリ上の状態・接続・稼働時間）を復元しないため、「巻き戻し可能」とは言えない
- start / stop は逆操作（stop / start）で稼働状態を戻せるなら `true` にできる。ただし停止で失われる内部状態が実質的な非可逆性を生む場合は `false` とする。判断はツールに委ねる

## restart 不要判定（差分なし）

判定の結果「restart 不要」となった場合の表現（→ ADR-0020 の帰結）。

- item は `status: "success"` とする。判定処理が正常に完走した結果であり、`skipped`（前段の失敗による未実行）にしない
- 必要なら背景を `info`（例: `{ "action": "restart", "needed": false }`）や `W_` 警告で運ぶ
- changes には列挙しない。差分の無い項目（noop）を changes に含めてはならない（→ `spec/v1/spec.md` §4）

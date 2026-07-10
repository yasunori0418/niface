#!/usr/bin/env python3
"""testdata の valid/ が schema + リントに適合し invalid/ が不適合であることを検証する。"""
import json, sys, glob, os
import jsonschema


def lint_envelope(doc):
    """schema で表現しきれない MUST を検査し、違反メッセージのリストを返す(§2 status / §5 参照・一意性)。"""
    findings = []
    results = doc.get("results", [])
    # subject.name の 1 エンベロープ内一意性(producer MUST)
    seen_names = set()
    for r in results:
        name = r.get("subject", {}).get("name")
        if name in seen_names:
            findings.append(f"subject.name が 1 エンベロープ内で重複: {name!r}")
        seen_names.add(name)
    for idx, r in enumerate(results):
        result = r.get("result", {})
        items = result.get("items", [])
        item_ids = [i.get("id") for i in items]
        # item.id の 1 result 内一意性(§5)
        seen_ids = set()
        for iid in item_ids:
            if iid in seen_ids:
                findings.append(f"results[{idx}] の item.id が result 内で重複: {iid!r}")
            seen_ids.add(iid)
        # changes[].itemId の同一 result 内参照整合(§5)
        id_set = set(item_ids)
        for c in result.get("changes", []):
            ref = c.get("itemId")
            if ref not in id_set:
                findings.append(f"results[{idx}] の changes.itemId が同一 result の item を指さない: {ref!r}")
    return findings


def main():
    schema_path, testdata_dir = sys.argv[1], sys.argv[2]
    schema = json.load(open(schema_path))
    v = jsonschema.Draft202012Validator(schema)
    ok = True
    for f in sorted(glob.glob(os.path.join(testdata_dir, "valid", "*.json"))):
        doc = json.load(open(f))
        errs = list(v.iter_errors(doc))
        lint = lint_envelope(doc)
        if errs:
            print(f"FAIL {f}: schema: {errs[0].message}"); ok = False
        elif lint:
            print(f"FAIL {f}: lint: {lint[0]}"); ok = False
        else:
            print(f"PASS {f}")
    for f in sorted(glob.glob(os.path.join(testdata_dir, "invalid", "*.json"))):
        doc = json.load(open(f))
        errs = list(v.iter_errors(doc))
        lint = lint_envelope(doc)
        if errs:
            print(f"PASS {f} (invalid として拒否: schema)")
        elif lint:
            print(f"PASS {f} (invalid として拒否: lint: {lint[0]})")
        else:
            print(f"FAIL {f}: 不正データが schema/lint を通過"); ok = False
    sys.exit(0 if ok else 1)


if __name__ == "__main__":
    main()

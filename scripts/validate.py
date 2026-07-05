#!/usr/bin/env python3
"""testdata の valid/ が schema に適合し invalid/ が不適合であることを検証する。"""
import json, sys, glob, os
import jsonschema

schema_path, testdata_dir = sys.argv[1], sys.argv[2]
schema = json.load(open(schema_path))
v = jsonschema.Draft202012Validator(schema)
ok = True
for f in sorted(glob.glob(os.path.join(testdata_dir, "valid", "*.json"))):
    errs = list(v.iter_errors(json.load(open(f))))
    if errs:
        print(f"FAIL {f}: {errs[0].message}"); ok = False
    else:
        print(f"PASS {f}")
for f in sorted(glob.glob(os.path.join(testdata_dir, "invalid", "*.json"))):
    errs = list(v.iter_errors(json.load(open(f))))
    if errs:
        print(f"PASS {f} (invalid として拒否)")
    else:
        print(f"FAIL {f}: 不正データが schema を通過"); ok = False
sys.exit(0 if ok else 1)

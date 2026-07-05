# niface

*Read this in [Japanese (日本語)](README.ja.md).*

**n**-tools **i**nter**face** — a shared JSON specification that lets the Nix-based system-management tools (nput / nboot / nwrap / nherd / nshadow / ncompose) talk to each other over stdout / stdin.

In a UNIX-philosophy ecosystem where each tool has a single responsibility and they compose through pipes to form a distribution, niface defines "the bore of the plumbing." Any tool, written in any language, can join a pipeline as long as it conforms to the spec.

This repository has two roles:

1. **The niface specification**: the normative spec, JSON Schema, conformance test data, and reference implementations.
2. **The ecosystem's central documentation**: the overall vision for the n-prefixed tools and the distribution plan ([docs/ecosystem/](docs/ecosystem/)).

> **Status: draft (specVersion 1)** — the initial design questions are settled. We are now in the phase of stacking backward-compatible changes (adding fields) driven by feedback from real tools.

## Documentation

The design and specification documents are written in Japanese.

Specification:

| Document | Contents |
|------|------|
| [docs/concept.md](docs/concept.md) | Why it exists, its principles, and its non-goals |
| [docs/design.md](docs/design.md) | All design decisions (19 items) and their rationale |
| [spec/v1/spec.md](spec/v1/spec.md) | **Normative** (the specification, in RFC 2119 vocabulary) |

Ecosystem:

| Document | Contents |
|------|------|
| [docs/ecosystem/overview.md](docs/ecosystem/overview.md) | Master index: the north star, tool list, and design principles |
| [docs/ecosystem/distro-plan.md](docs/ecosystem/distro-plan.md) | The distribution vision, each tool's responsibility, and milestones |

## Repository layout

```
niface/
├── docs/             # The spec's concept and design
│   └── ecosystem/    # Ecosystem central docs (vision, naming, master index)
├── spec/v1/          # The normative specification
├── schema/v1/        # JSON Schema (the machine-readable source of truth)
├── testdata/v1/      # Conformance test data
│   ├── valid/        #   samples that must pass the schema
│   ├── invalid/      #   samples the schema must reject
│   └── id-vectors.json  # identity → expected id table (the key to cross-language compatibility)
├── go/               # Go reference implementation (envelope types + id derivation)
├── nix/              # Nix implementation (id derivation)
├── scripts/          # Validation scripts
├── dev/              # Development environment (devShell + mattpocock/skills placement)
└── flake.nix
```

## The spec in 30 seconds

Every tool writes a single JSON document (the envelope) — and nothing else — to stdout:

```jsonc
{
  "specVersion": 1,
  "tool": { "name": "nput", "version": "0.9.0" },
  "command": "apply",
  "status": "success",            // "success" | "error" (only two values)
  "dryRun": false,                // a plan uses the same schema as apply
  "startedAt": "...", "finishedAt": "...",
  "errors": [],                   // only whole-run errors not tied to an item
  "result": {
    "items":   [ ... ],           // per-unit execution results (one failed → error)
    "changes": [ ... ],           // declared diffs (each always carries `reversible`)
    "info": { }                   // tool-specific fields live only here (same inside items/changes)
  }
}
```

The item id is derived mechanically as `sha256(JCS(identity))`. A tool decides only the contents of `identity` (`{kind, key}`); the spec absorbs the problems of format, escaping, and uniqueness.

## Conformance

A tool is niface-conformant when:

1. **Schema validation**: its output passes `schema/v1/envelope.schema.json`.
2. **id-vectors**: its id-derivation implementation matches the expected values for every vector in `testdata/v1/id-vectors.json`.
3. **standalone**: input comes only from stdin JSON or explicit arguments; it never implicitly discovers state or configuration (see spec §8).

## Validation

```sh
nix flake check   # runs id-vectors (Nix impl) + schema (testdata) + go test
```

Individually:

```sh
# Schema validation
python3 scripts/validate.py schema/v1/envelope.schema.json testdata/v1

# The Go implementation's vector test
cd go && go test ./...
```

> **Note**: the Go / Nix implementations were authored in an environment without the Go and Nix toolchains, so they are **unrun**. Confirm that `nix flake check` passes before the first push. The `id-vectors.json` values themselves are real values computed by the Python implementation (a JCS subset).

## Development environment

Copy the template `.envrc` and allow direnv. It loads the flake under `dev/` (`use flake ./dev`), which installs the dev tooling and places mattpocock/skills into `.claude/skills/` via nput's project mode.

```sh
cp example.envrc .envrc && direnv allow    # or: nix develop ./dev
```

## Related

- [nput](https://github.com/yasunori0418/nput) — the placement primitive (the ecosystem's origin; active).
- The other tools (nboot / nwrap / nherd / nshadow / ncompose) are planned. See [docs/ecosystem/overview.md](docs/ecosystem/overview.md).

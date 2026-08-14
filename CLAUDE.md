# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`qdiimpl` is a single-purpose Go code generator CLI. Given an interface, it emits a
struct implementing it where each method dispatches to caller-supplied callbacks
(`WithMETHOD(...)` functional options). Intended for debugging, not production.

## Commands

```shell
go build ./...
go test ./...                                   # only internal/util has tests
go test ./internal/util/ -run TestGetQualCode   # single test
go vet ./...
```

Run the generator from source against a directory:

```shell
cd sample/package && go run ../.. -type=Reader -type-package=io -overwrite=true -force-package-name=main
```

Each `sample/*/main.go` has the exact flags it was generated with in a comment at the
top of the file — use those to regenerate after changing `gen()`.

Tasks (`Taskfile.yml`): `task install`, `task current-version`,
`task release-version VERSION=vX.Y.Z` (tags both `vX.Y.Z` and `qdii/vX.Y.Z`, requires a
clean tree).

## Architecture

`main.go` → `cmd.Main()`. Essentially all logic is in **`cmd/qdiimpl.go`**:

1. Flags are package-level `flag.*` globals, read directly inside `gen()` — there is no
   config struct, so `gen()` reads generator options as `*flagName` throughout.
2. `run()` loads the target package with `golang.org/x/tools/go/packages` via
   `util.PkgInfoFromPath`, looks the type up in the package scope, and verifies it is an
   interface. `-all-interfaces` iterates every exported named interface in scope instead.
3. `gen()` builds the whole output file with the `dave/jennifer` DSL (dot-imported as
   `.`), walking `iface.Methods()`.

**`internal/util`** holds the reusable pieces:

- `jenutil.go` — `GetQualCode` converts any `go/types.Type` into a jennifer statement
  (basics, arrays, slices, pointers, tuples, interfaces, maps, chans, named, aliases,
  type params, signatures). This is the heart of type rendering; extend it here rather
  than inline in `gen()`. Also `GetSignatureParamQualCode`/`GetSignatureParamCallCode`
  for variadics and `AddTypeParamsList`/`AddTypeList` for generics.
- `typesutil.go` — parameter naming (`ParamName` for unnamed params → `p0`, `p1`;
  `MethodParamName` renames a param called `d` to `d1` because `d` is the receiver).
- `util.go` — `GetUniqueName` appends a `QDII` suffix until a generated identifier no
  longer collides with an interface method name.

**`qdii/`** is a separate nested module containing only the `Context` interface
(`MethodName`, `ExecCount`, `CallerFunc`, `CallerFile`, `CallerLine`, `NotSupported`).
Generated `TYPEContext` types satisfy it, so consumers can accept `qdii.Context` without
depending on the generator. Keep it dependency-free.

## Shape of the generated code

For interface `T`, `gen()` emits: `TContext` (unexported fields + accessor methods, plus
`Data` when `-data-type` is set), the impl struct `T` holding one
`implMETHOD []func(qdCtx *TContext, ...) (...)` slice per method, `TOption`, `NewT`, one
method per interface method, the private helpers (`getCallerFuncName`, `addCallMethod`,
`createContext`, `methodNotImplemented`, optionally `callLogger`), and the `WithXXX`
options.

Method dispatch order — this is the core semantic:

1. Iterate the registered callbacks for that method in registration order. Each
   `WithMETHOD` call **appends**, so multiple callbacks can be registered.
2. If a callback calls `qdCtx.NotSupported()`, its return values are discarded and the
   next callback is tried.
3. If no callback handled it, delegate to the `fallback` (a value, or a
   `func() (T, error)` when `-fallback-func` is set).
4. Otherwise `panic(d.methodNotImplemented(...))`, which routes through the
   `WithOnMethodNotImplemented` hook first if one is set.

## Invariants when editing `gen()`

- Any new struct field or option name must go through `util.GetUniqueName` — interface
  methods share the identifier namespace with generated fields (this is why `fallback`,
  `Data`, `onMethodNotImplemented`, `onCallLogger` all have `*ParamName` variables).
- Generic interfaces: *uses* of the impl type need `.TypesFunc(codeObjectTypes)`,
  *declarations* need `codeObjectTypesWithType`. The `var _ T = (*T)(nil)` assertion is
  skipped entirely when the interface has type params.
- Never render parameter types with `GetQualCode` directly for method signatures — use
  `GetSignatureParamQualCode` so variadics keep their `...`.
- `-export-type` and `-name-prefix` interact: the prefix is upper-cased for exported
  names and lower-cased for the unexported struct name, unless the caller already
  supplied a mixed-case prefix.

## Testing

Tests use `gotest.tools/v3/assert`, not testify. `internal/util/base_test.go` provides
the harness: build a `[]tc` of `{desc, code|codeFn, expect}`, where `expect` is Go source
that gets `format.Source`-normalized and compared against jennifer's `%#v` render, then
pass it to `runTestCases`. New type-rendering support belongs there as a new `tc` entry
in `jenutil_test.go`.

## Caveats

- The checked-in `sample/*/*_qdii.go` files are **stale** — generated by older versions
  and not part of the build verification. Regenerate before trusting them as reference.
- `qdii/go.mod` declares module path `github.com/rrgmc/qdiimpl`, the same as the root
  module; it likely should be `.../qdiimpl/qdii`.
- CI (`.github/workflows/go.yml`) pins Go 1.23 while `go.mod` requires Go 1.26.
- The README flag list is out of date — it omits `-all-interfaces`, `-skip-interfaces`,
  `-call-logger`, `-fallback-func`, `-data-type-self`, `-option-name-prefix-self`, and
  `-no-format`. `cmd/qdiimpl.go` is the source of truth.

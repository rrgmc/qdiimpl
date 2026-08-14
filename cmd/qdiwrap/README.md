# qdiwrap (POC)

A proof-of-concept alternative to `qdiimpl` that builds its output from `text/template`
templates driven by [`github.com/hexdigest/gowrap`](https://github.com/hexdigest/gowrap)
instead of the `dave/jennifer` AST DSL.

It implements only the two core features of `qdiimpl`:

1. Overriding each interface method individually with a `WithMETHOD(...)` functional option.
2. Setting a `fallback` implementation to delegate to when a method was not overridden.

There is no `TYPEContext`, no `ExecCount`, no caller introspection, no `NotSupported()`
retry chain, no `-data-type` and no `-call-logger`. If a method has neither an override nor
a fallback, it panics.

`gowrap` does the heavy lifting `internal/util/jenutil.go` does for `qdiimpl`: parsing the
interface, rendering parameter and result types, handling generics and variadics, and
resolving imports through `goimports`. As a result `cmd/qdiwrap/qdiwrap.go` is ~130 lines
against `cmd/qdiimpl.go`'s 788, and all of the generated shape lives in
`templates/body.tmpl`.

## Usage

```shell
cd sample/qdiwrap
go run ../../qdiwrap -type=ReadWriter -type-package=io -overwrite=true -force-package-name=main
```

| flag | default | notes |
|---|---|---|
| `-type` | | type name; must be set |
| `-type-package` | `""` | type package path if not the current directory |
| `-force-package-name` | `""` | force generated package name |
| `-name-prefix` | `""` | interface name prefix (needed for same-package generation) |
| `-name-suffix` | `""` | interface name suffix |
| `-option-name-prefix` | `""` | option name prefix (`WithQWRead`) |
| `-output` | `""` | output file name; default `srcdir/<type>_qdiw.go` |
| `-overwrite` | `false` | overwrite file if exists |

## Generated shape

See `sample/qdiwrap/readwriter_qdiw.go`. For each interface method `M` the output has an
`implM func(...)` field, an `M(...)` method that tries `implM`, then `fallback.M(...)`,
then panics, and a `WithM(...)` option that sets `implM`.

`New<T>` returns the concrete `*T` rather than the interface, so callers can still reach
the struct; it satisfies the interface either way (`var _ T = (*T)(nil)` is emitted, except
for generic interfaces).

## Differences from qdiimpl, and gowrap-imposed limits

These are part of what this POC is measuring.

- **Single override, last wins.** `implM` is a plain func field, not a slice. Without a
  context there is no `NotSupported()` signal, so an append-list could never advance past
  the first callback.
- **The receiver is `_d`**, following gowrap's own bundled templates. This sidesteps
  `util.MethodParamName`'s `d` → `d1` collision workaround entirely.
- **Unexported methods are rejected when generating cross-package** (gowrap's
  `errUnexportedMethod`). Same-package generation is fine — `sample/complex`'s
  `internal() bool` works with `-name-prefix`.
- **Empty interfaces are rejected** (gowrap's `errEmptyInterface`).
- **Parameter names differ.** gowrap auto-names anonymous parameters by type prefix (`b1`
  for `[]byte`, `s1`, `ch1`, `sp1`) rather than `qdiimpl`'s `p0`/`p1`, and forces `ctx` for
  a leading `context.Context` and `err` for a trailing `error`. Output will not match the
  existing `*_qdii.go` samples byte-for-byte.
- **No `GetUniqueName` guard.** An interface declaring an unexported method literally named
  `fallback` would collide with the struct field. `qdiimpl` handles this;
  this POC does not.
- **No `-tags`.** gowrap's `pkg.Load` hardcodes its `packages.Config` with no `BuildFlags`,
  so build tags cannot be plumbed through `SourcePackage`. Working around it would mean
  loading the package ourselves and passing `Options.SourcePackageInstance`.
- **No `-all-interfaces` / `-skip-interfaces`**, and no `-sync`, `-export-type`,
  `-no-format` or `-same-package`. `Generate` always runs `goimports`, so `-no-format` has
  no equivalent.

The source package is aliased to its own name (`io`) rather than gowrap's default
`_sourceIo`, so the doc links in the generated comments resolve.

# Performance

Parser performance matters most for normal parsing.
Help, version, completion, docs, and config-example generation
are usually control-flow exits or build-time operations.

Optimize real application paths before optimizing one-shot output paths.

## Reusing Parsers

A reused parser avoids repeated reflection scanning and metadata setup.
This is useful in tests, benchmarks, embedded tools,
and long-running processes that parse many argument lists.

```go
parser := flags.NewParser(&opts, flags.Default&^flags.PrintErrors)
for _, args := range cases {
  _, err := parser.ParseArgs(args)
  _ = err
}
```

Be careful with reused parsers and mutable target structs.
Parsing writes into the same values.
Reset state intentionally between parses when needed.

## New Parser Per Run

A normal CLI process usually builds one parser,
parses once, executes, and exits.

For that path, clarity matters more than micro-optimizing setup.
Avoid global parser caches unless the application
has a real repeated-parse workload.

## Dirty Metadata Validation

The parser tracks when metadata changes require duplicate validation
to run again.
This avoids repeated validation work for reused parser instances
while still protecting programmatic mutation paths.

Programmatic changes through setters, rebuilds, tag remapping,
or configurators can mark metadata dirty.
The next parse or validation pass refreshes checks.

This is useful because parser metadata can change after construction.
Skipping validation forever would be incorrect.
Running full validation every time would penalize reused parsers.

## Help and Docs

`WriteHelp` is optimized enough for runtime use,
but most applications print help and exit.
Caching help output in application code is rarely useful.

`WriteDoc` is intended for generated documentation.
It can do more work because it is usually run manually,
in CI, or through helper commands.

Do not add persistent caches for one-shot docs unless measurement
shows a real problem in a real workflow.

## Completion

Completion can run often while the user presses tab.
Startup time matters.

Keep completion-safe startup light:

* avoid network calls before parsing;
* avoid slow config loading when not needed for completion;
* avoid logging noise during completion;
* keep custom completers fast.

If completion needs dynamic data,
cache it in application code with clear invalidation.

## Environment and Filesystem Checks

Validators such as `validate-existing-file`, `validate-readable`,
and `validate-writable` touch the filesystem.
This is correct for validation, but it is not free.

Use these validators for values that truly need filesystem checks
at parse time.
Use application validation when checks depend on later processing anyway.

## Benchmarks

Project benchmarks cover parser setup, reused parsing, help, man/docs,
INI, completion script generation, version output, localization,
and catalog coverage.

Use benchmark changes as signals, not as automatic proof of a bad patch.
A slower benchmark may be acceptable if the path is one-shot
and the feature adds useful behavior.

Always look at allocations and call trees for hot paths before adding caches.

## What Not to Optimize

Avoid optimizing:

* help output that immediately exits;
* version output that immediately exits;
* docs rendering that runs in CI or release tooling;
* config example output that users call occasionally;
* code paths made slower only by synthetic benchmark setup.

Do optimize:

* repeated `ParseArgs` use;
* completion startup;
* reflection scanning that happens more than needed;
* allocations in hot parse loops;
* unnecessary filesystem work in common parse paths.

## Optimization Rules

Measure before changing architecture. Prefer small localized improvements.
Remove benchmark-only complexity when
it makes normal code harder to reason about.

A CLI parser should be fast, but predictable behavior
and clear contracts are usually more valuable
than saving microseconds in one-shot paths.

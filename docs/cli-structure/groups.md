# Groups

Groups organize related options in parser structure,
help output, generated docs, environment namespaces, and INI sections.

Use groups when options share a topic, a prefix,
or a configuration file section.
Do not use groups only to make the Go struct look nested.
The nesting should help users understand the CLI.

## Basic Group

A nested struct tagged with `group` becomes an option group.

```go
type Options struct {
  Logging struct {
    Level string `long:"level" default:"info" choices:"debug;info;warn;error"`
    JSON  bool   `long:"json"`
  } `group:"Logging" description:"Logging options"`
}
```

The group heading is used in help and generated docs.  
The child options remain normal options.

## Namespaces

`namespace` prefixes child long option names.

```go
type Options struct {
  DB struct {
    Host string `long:"host"`
    Port int    `long:"port" default:"5432"`
  } `group:"Database" namespace:"db"`
}
```

The command line uses `--db.host` and `--db.port`.  
The delimiter is controlled by `Parser.NamespaceDelimiter`.  
The default delimiter is `.`.

Use namespaces when the prefix is part of the public CLI design.
Avoid namespaces when they only repeat the group title
without reducing ambiguity.

## Environment Namespaces

`env-namespace` prefixes child environment keys.

```go
type Options struct {
  DB struct {
    Host string `long:"host" env:"HOST"`
  } `group:"Database" env-namespace:"DB"`
}
```

With `parser.SetEnvPrefix("APP")`, the final key is `APP_DB_HOST`.
The delimiter is controlled by `Parser.EnvNamespaceDelimiter`.
The default delimiter is `_`.

Use env namespaces to keep environment variables predictable in large CLIs.

## INI Sections

Groups create INI sections when INI read/write is used.

`ini-group` provides a stable section token.
Use it when display names are localized or may change.

```go
type Options struct {
  Cache struct {
    Dir string `long:"dir" ini-name:"dir"`
  } `group:"Cache Settings" ini-group:"cache"`
}
```

Do not let localized group names become config-file section names.
A config file should not change shape when locale changes.

## Nested Groups

Groups can be nested.
Nested groups can combine namespaces and INI section tokens.

```go
type Options struct {
  Cloud struct {
    Auth struct {
      Token string `long:"token" env:"TOKEN" no-ini:"true"`
    } `group:"Auth" namespace:"auth" env-namespace:"AUTH" ini-group:"auth"`
  } `group:"Cloud" namespace:"cloud" env-namespace:"CLOUD" ini-group:"cloud"`
}
```

* Nested long names become hierarchical.
* Nested env names become hierarchical.
* Nested INI sections become hierarchical.

Keep nesting shallow.
Deep option trees are harder to type and harder to document.

## Hidden Groups

`hidden:"true"` hides the group and its options from help,
completion, and generated docs. The options remain parseable.

Use hidden groups for compatibility or internal switches.
Do not use them as a security boundary.

## Immediate Groups

`immediate:"true"` marks all options in the group subtree as immediate.
When an immediate option is set,
required checks and command execution are skipped.

This is useful for control-flow groups such as help,
version, completion, or docs.
It should not be used for ordinary business options.

## Programmatic Groups

Groups can also be added with `Parser.AddGroup`.
Use this for plugins, dynamic modules, or generated option sets.

Struct tags are usually clearer for static applications.
Programmatic groups should still have stable names, descriptions,
and INI section tokens if they are public.

## Runtime Setters

`Group` exposes setters for metadata such as descriptions,
namespaces, INI name, hidden state, and immediate state.

Use setters when values are discovered at runtime.
Use tags when values are part of the stable CLI contract.

## Group Design Rules

Group by user task, not by internal package name.

A good group heading helps the user decide whether to read the options inside.
If the heading is vague, the group is probably not useful.

Use namespaces only when option names would otherwise collide or become unclear.
Short CLIs often read better without namespaced long options.

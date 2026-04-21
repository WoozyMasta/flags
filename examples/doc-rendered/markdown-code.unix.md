# golden-doc

## NAME

**golden-doc** - Golden doc parser

## SYNOPSIS

`golden-doc [OPTIONS]`

## DESCRIPTION

Long description for golden tests.
Includes options, groups and commands.

## OPTIONS

### Application Options

```text
-v, --verbose - Enable verbose output
--config [def: config.yaml] [env: MY_APP_APP_CONFIG] [req] - Path to config
--mode [def: fast] [choices: fast, safe] - Execution mode
--tag [def: api] - Tag filter
--header [def: x-env:dev] [kv: :] - HTTP headers
--level [=info] [optional: info] - Log level
```

### Database Options

```text
--db.host [def: 127.0.0.1] [env: MY_APP_DB_HOST] - Database host
--db.port [def: 5432] [env: MY_APP_DB_PORT] - Database port
```

## COMMANDS

### run

Run command

Execute deployment workflow.

**Usage:** `golden-doc [OPTIONS] run [run-OPTIONS]`

#### Run command

```text
--force - Force execution
--plan - Show execution plan only
```

#### Arguments

```text
target - Deployment target
```

### status

Show status

Read and print current status.

**Usage:** `golden-doc [OPTIONS] status [status-OPTIONS]`

#### Show status

```text
--json - JSON output
```

## ARGUMENTS

```text
input - Input resource
```

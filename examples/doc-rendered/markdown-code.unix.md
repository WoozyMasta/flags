<!-- markdownlint-disable MD013 MD024 MD036 -->
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
--config [default: config.yaml] [environment: MY_APP_APP_CONFIG] [required] - Path to config
--mode [default: fast] [choices: fast, safe] - Execution mode
--tag [default: api] - Tag filter
--header [default: x-env:dev] [delimiter: :] - HTTP headers
--level [=info] [optional value: info] - Log level
```

### Database Options

```text
--db.host [default: 127.0.0.1] [environment: MY_APP_DB_HOST] - Database host
--db.port [default: 5432] [environment: MY_APP_DB_PORT] - Database port
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

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

|Option|Description|Default|Environment|Required|
|---|---|---|---|---|
|`-v`, `--verbose`|Enable verbose output|||no|
|`--config`|Path to config|config.yaml|MY_APP_APP_CONFIG|yes|
|`--mode`|Execution mode; choices: `fast, safe`|fast||no|
|`--tag`|Tag filter|api||no|
|`--header`|HTTP headers|x-env:dev|; kv delim: `:`|no|
|`--level [=info]`|Log level; optional: `info`|||no|

### Database Options

|Option|Description|Default|Environment|Required|
|---|---|---|---|---|
|`--db.host`|Database host|127.0.0.1|MY_APP_DB_HOST|no|
|`--db.port`|Database port|5432|MY_APP_DB_PORT|no|

## COMMANDS

### run

Run command

Execute deployment workflow.

**Usage:** `golden-doc [OPTIONS] run [run-OPTIONS]`

#### Run command

|Option|Description|Required|
|---|---|---|
|`--force`|Force execution|no|
|`--plan`|Show execution plan only|no|

#### Arguments

|Name|Description|Required|
|---|---|---|
|`target`|Deployment target|no|

### status

Show status

Read and print current status.

**Usage:** `golden-doc [OPTIONS] status [status-OPTIONS]`

#### Show status

|Option|Description|Required|
|---|---|---|
|`--json`|JSON output|no|

## ARGUMENTS

|Name|Description|Required|
|---|---|---|
|`input`|Input resource|no|

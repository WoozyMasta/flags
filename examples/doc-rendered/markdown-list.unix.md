<!-- markdownlint-disable MD013 MD024 MD036 -->
# golden-doc

## NAME

**golden-doc** * Golden doc parser

## SYNOPSIS

`golden-doc [OPTIONS]`

## DESCRIPTION

Long description for golden tests.
Includes options, groups and commands.

## OPTIONS

### Application Options

Main options

* `-v`, `--verbose` -
  Enable verbose output
* `--config` -
  Path to config
  * Required: `yes`
  * Defaults: `config.yaml`
  * Environment: `MY_APP_APP_CONFIG`
* `--mode` -
  Execution mode
  * Defaults: `fast`
  * Choices: `fast, safe`
* `--tag` -
  Tag filter
  * Defaults: `api`
* `--header` -
  HTTP headers
  * Defaults: `x-env:dev`
  * delimiter: `:`
* `--level [=info]` -
  Log level
  * Optional value: `info`

### Database Options

* `--db.host` -
  Database host
  * Defaults: `127.0.0.1`
  * Environment: `MY_APP_DB_HOST`
* `--db.port` -
  Database port
  * Defaults: `5432`
  * Environment: `MY_APP_DB_PORT`

## COMMANDS

### run

Run command

Execute deployment workflow.

**Usage:** `golden-doc [OPTIONS] run [run-OPTIONS]`

#### Run command

Execute deployment workflow.

* `--force` -
  Force execution
* `--plan` -
  Show execution plan only

#### Arguments

* `target`
  * Description: Deployment target

### status

Show status

Read and print current status.

**Usage:** `golden-doc [OPTIONS] status [status-OPTIONS]`

#### Show status

Read and print current status.

* `--json` -
  JSON output

## ARGUMENTS

* `input`
  * Description: Input resource

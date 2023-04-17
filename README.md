# JIRA-EXPORTER

This is a simple tool for exporting JIRA JQL queries to CSV or JSON files.

## Requirements

* Go 1.20
* go-task 3.2 (https://taskfile.dev/installation)

## Preparation

```bash
cp .env.example .env
```

Edit the `.env` file to match your JIRA instance.

## Usage

```bash
jira-export --help
Error loading .env file
Export Jira issues to CSV and JSON

Usage:
  jira-export [flags]

Flags:
  -h, --help              help for jira-export
  -j, --jql string        JQL query
  -m, --max-results int   Max results (default 100)
  -o, --output string     Output directory (default "dist/jira/results")
  -t, --token string      Jira token
  -r, --url string        Jira URL
  -u, --username string   Jira username
```

Using the Taskfile.yaml
```bash
task run
```

Or using the docker container

```bash
task docker::run
```

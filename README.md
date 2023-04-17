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
task run
```

Or using the docker container

```bash
task docker::run
```

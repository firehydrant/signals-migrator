# Signals Migrator

This can be used to import resources from legacy alerting providers into Signals.

## Supported providers

### Alerting 

- PagerDuty
- VictorOps
- OpsGenie

## Usage

### Generate Signals resources from external alerting providers

Ensure these environment variables are set:

- `FIREHYDRANT_API_KEY`
- `PROVIDER` e.g. 'PagerDuty'
- `PROVIDER_API_KEY`
- `PROVIDER_APP_ID` (optional, not all providers require this)

Afterwards, run `signals-migrator import` (or `go run . import` for development version), which will generate `output/[PROVIDER]_to_fh_signals.tf` file.

During the process, we will attempt to match users by email to existing users in FireHydrant. For users without a match, we will ask you to decide on whether to skip the user or manually match them to existing user.

> [!IMPORTANT]  
> If you are using Single Sign-On (SSO) for FireHydrant, users of your organization may need to log in to FireHydrant _at least once_ before running the migration tool.

On the other hand, we can't reliably match teams ourselves as they have wide variance of identification. As such, we will ask you to select from three options:

1. Skip the team
1. Create a new team
1. Match to an existing team

Afterwards, the tool will generate the mapping appropriately, handling de-duplication and merging as necessary.

## Feature roadmap

| | PagerDuty | Opsgenie | VictorOps |
| --- | --- | --- | --- |
| Import users | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| Import teams and members | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| Import escalation policies | :white_check_mark: | :white_check_mark: | :x: |
| Import scheduling strategy | :white_check_mark: | :white_check_mark: | :x: |


- [ ] Getting transposer URLs (e.g. Datadog) to the team data resource or a Signals ingest URL data resource
- [ ] Auto-run `terraform apply` for users who would not manage their organization with Terraform after importing

## Developing

A devcontainer setup has been prepared to be used in VS Code. Run `direnv allow` to auto-load `.env` file.

If alternative method is preferred, you will need:

- Go compiler
- Install tools in `./deps.sh`

Also recommended to smooth out development outside of devcontainer:

- [direnv](https://direnv.net/) for autoloading `.env` file, also automatically adds `./bin` to `$PATH`.
- [just](https://just.systems/) for running tasks defined in `Justfile`.

Most commands in `Justfile` can be run out in regular bash too.

# PagerDuty to Signals

We support importing from PagerDuty to Signals. To get started, please set up the following environment variables:

```shell
export FIREHYDRANT_API_KEY=your-firehydrant-api-key
export PROVIDER=PagerDuty
export PROVIDER_API_KEY=your-pagerduty-api-key
```

> [!NOTE]
> Read-only key for PagerDuty should be sufficient.

Afterwards, run `signals-migrator import` and follow the prompts.

## Known limitations

- While we support importing "PagerDuty Service" as "FireHydrant Team", we still require the Teams API to be accessible. If your account does not have access to the Teams API, please see [#27](https://github.com/firehydrant/signals-migrator/issues/27) and let us know what error you encountered.
- Inactive users will have "warning" lines.

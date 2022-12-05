# portainer-github-webhook

A simple webhook to filter and forward requests to portainer for a specific branch.

## Config Example

The webhook url would be `http://yourhost:3473?branch=master&uuid=abc-def-ghi-hjk` where `uuid` is your portainer webhook uuid after `/api/stacks/webhooks/` and branch is the desired branch to cause invocations.

```yaml
# config.yaml

portainer_url: http://sub.domain.tld
secret_key: abcdefghijklmnopqrstuvwxyz1234567890  # GitHub webhook secret key
```

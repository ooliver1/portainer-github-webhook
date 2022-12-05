# portainer-github-webhook

A simple webhook to filter and forward requests to portainer for a specific branch.

## Config Example

Content type is json.

The webhook url would be `http://yourhost:3473?branch=master&uuid=abc-def-ghi-hjk` where `uuid` is your portainer webhook uuid after `/api/stacks/webhooks/` and branch is the desired branch to cause invocations.

```properties
PORTAINER_URL=http://sub.domain.tld
SECRET_KEY=abcdefghijklmnopqrstuvwxyz1234567890  # GitHub webhook secret key
```

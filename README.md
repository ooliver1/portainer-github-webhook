# portainer-github-webhook

A simple webhook to filter and forward requests to portainer for a specific branch.

## Config Example

```yaml
# config.yaml

portainer_url: http://sub.domain.tld
secret_key: abcdefghijklmnopqrstuvwxyz1234567890  # GitHub webhook secret key
webhooks:
  - repo: user/repo
    branch: master
    uuid: abc-def-ghi-jkl-mno  # Portainer webhook uuid (after /api/stacks/webhooks/)
  - repo: org/repo
    branch: dev
    uuid: pqr-stu-vwx-yza-bcd
```

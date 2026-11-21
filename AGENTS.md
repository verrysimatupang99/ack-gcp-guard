# AGENTS.md

## Project
`ack-gcp-guard` — production cost & credential guard for Alibaba Cloud ACK and GCP GKE.

## How to run
- Go CLI: `go build -o bin/guard ./cmd/guard` then `./bin/guard --help`.
  Subcommands: `ram-scanner`, `gcp-scanner`, `billing-watcher`.
- Creds are env vars only (never hardcode):
  - Alibaba: `ALIBABA_CLOUD_ACCESS_KEY_ID`, `ALIBABA_CLOUD_ACCESS_KEY_SECRET`, optional `ALIBABA_CLOUD_REGION` (default cn-hangzhou).
  - GCP: `GOOGLE_APPLICATION_CREDENTIALS` (path to SA key JSON), `GCP_PROJECT_ID`, `GCP_BILLING_DATASET`, `GCP_BILLING_TABLE`.
  - Alerts: `SLACK_WEBHOOK_URL` / `DINGTALK_WEBHOOK_URL` (optional).
- Terraform: `cd terraform && terraform init && terraform plan`.

## Pitfalls
- Alibaba OpenAPI throttling: retry with exponential backoff is built into the RAM scanner (`isThrottle` + `withRetry`). Don't remove it.
- BigQuery billing export must be enabled in GCP before the billing watcher can query it (see `terraform/modules/gcp-billing`).
- Go module path is `github.com/si_juling/ack-gcp-guard`; update if the repo owner changes.

## Publish to GitHub (finish the last step)
Local git is already initialized (`main` branch) and fully committed; remote `origin` is set to `https://github.com/si_juling/ack-gcp-guard.git`. To push:

```bash
# 1. Create a public repo first (one time), e.g. via gh:
#    gh repo create ack-gcp-guard --public --source . --push
#    Or create it in the GitHub UI, then:
git remote set-url origin https://github.com/<your-username>/ack-gcp-guard.git
# 2. Push (use a PAT with `repo` scope when prompted):
git push -u origin main
```

If the GitHub username differs from `si_juling`, run `git remote set-url origin ...` with the correct URL before pushing.

## TODO (for Devin Bots)
- Parallelize the three workers (currently sequential subcommands).
- Add `--json` output + CI-friendly exit codes.
- Add unit tests with mocked SDK clients.

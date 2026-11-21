# Build `ack-gcp-guard` — production cost & credential guard for ACK (Alibaba Cloud) + GKE (GCP)

## What & why
A public, production-grade repo that guards against the two biggest cluster-ops leaks: **stale cloud credentials** (Alibaba RAM keys + GCP Service Account keys older than 90 days) and **cost/billing anomalies** (day-over-day billing spikes >20%, CrashLoopBackOff pods). Goal: a real, build-in-public repo to win over Poke at Cognition for Devin Bots credits.

## Stages

### 1. Repo foundation
- `LICENSE` (MIT)
- `README.md`: banner (ASCII/SVG-style), problem statement (RAM key + SA key sprawl, ACK/GKE CrashLoopBackOff, billing spikes, Alibaba OpenAPI throttling, GCP Billing BigQuery export), Mermaid architecture diagram, "Built with Devin Bots" badge, live demo section, "Follow progress on X @si_juling" link, quickstart, env-var docs (ALIBABA_CLOUD_ACCESS_KEY / GCP_PROJECT_ID), TODO list for Devin Bots (parallel execution).
- `AGENTS.md` (project notes for future sessions).

### 2. Terraform (`.tf` ~1.5, `alibabacloud` + `google` providers)
- `terraform/modules/alibaba-ram/`: RAM user, read-only key-scanning policy, STS role + trust policy.
- `terraform/modules/gcp-billing/`: enable Billing BigQuery export, SA to read the billing dataset.
- `terraform/main.tf`, `variables.tf`, `outputs.tf`, `versions.tf` wiring both modules.

### 3. Go CLI (Go 1.22, Cobra)
- `go.mod` + `go.sum` (via `go mod init` / `go mod tidy`).
- `cmd/guard/main.go`: Cobra root command with 3 worker subcommands.
- `internal/scanner/ram.go`: `ram-scanner` — list RAM users via `alibabacloud-go/ram`, flag keys >90 days, with OpenAPI throttling retry (backoff).
- `internal/scanner/gcp.go`: `gcp-scanner` — list SA keys via `google.golang.org/api/iam`, flag >90 days.
- `internal/billing/watcher.go`: `billing-watcher` — query BigQuery billing table, alert when day-over-day >20%.
- `pkg/watcher/k8s_watcher.go`: watch CrashLoopBackOff in ACK/GKE, notify via Slack/DingTalk webhook.
- No hardcoded creds — env vars only.

### 4. CI
- `.github/workflows/ci.yml`: Go build + test, `terraform fmt -check` + `terraform validate`.

### 5. Publish
- `git init`, initial commit.
- Create public repo `github.com/<your-username>/ack-gcp-guard` and push (requires your GitHub username + Personal Access Token — see question).
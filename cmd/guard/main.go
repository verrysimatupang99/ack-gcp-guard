// Command guard is the ack-gcp-guard CLI. It exposes three worker
// subcommands (ram-scanner, gcp-scanner, billing-watcher) plus the
// k8s-watcher.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/si_juling/ack-gcp-guard/internal/billing"
	"github.com/si_juling/ack-gcp-guard/internal/scanner"
	"github.com/si_juling/ack-gcp-guard/pkg/watcher"
)

var version = "dev"

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "guard",
		Short: "ack-gcp-guard — cost & credential guard for ACK and GKE",
		Long: `ack-gcp-guard is a production cost & credential guard for Alibaba Cloud ACK
and Google Cloud GKE. It scans for stale cloud keys (RAM / service account),
watches CrashLoopBackOff pods, and alerts on billing spikes.

Workers:
  ram-scanner       flag Alibaba RAM access keys older than N days
  gcp-scanner       flag GCP service account keys older than N days
  billing-watcher   alert when day-over-day spend exceeds a threshold
  k8s-watcher       watch CrashLoopBackOff pods in ACK/GKE`,
		Version: version,
	}

	root.AddCommand(scanner.NewRAMScannerCmd())
	root.AddCommand(scanner.NewGCPScannerCmd())
	root.AddCommand(billing.NewBillingWatcherCmd())
	root.AddCommand(watcher.NewK8sWatcherCmd())

	return root
}

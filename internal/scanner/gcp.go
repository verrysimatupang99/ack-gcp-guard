package scanner

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2/google"
	iam "google.golang.org/api/iam/v1"
	"google.golang.org/api/option"

	"github.com/si_juling/ack-gcp-guard/pkg/notify"
)

// NewGCPScannerCmd returns the gcp-scanner worker command.
func NewGCPScannerCmd() *cobra.Command {
	var thresholdDays int
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "gcp-scanner",
		Short: "Flag GCP service account keys older than N days",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := os.Getenv("GCP_PROJECT_ID")
			if projectID == "" {
				return fmt.Errorf("GCP_PROJECT_ID must be set")
			}

			svc, err := newIAMService(cmd.Context())
			if err != nil {
				return err
			}

			return scanSAKeys(cmd.Context(), svc, projectID, thresholdDays, dryRun)
		},
	}

	cmd.Flags().IntVar(&thresholdDays, "threshold-days", 90, "flag keys older than this many days")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print findings without alerting")

	return cmd
}

func newIAMService(ctx context.Context) (*iam.Service, error) {
	creds, err := google.FindDefaultCredentials(ctx, iam.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("find default credentials: %w", err)
	}
	return iam.NewService(ctx, option.WithCredentials(creds))
}

func scanSAKeys(ctx context.Context, svc *iam.Service, projectID string, thresholdDays int, dryRun bool) error {
	cutoff := time.Now().AddDate(0, 0, -thresholdDays)

	sas, err := svc.Projects.ServiceAccounts.List("projects/" + projectID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("list service accounts: %w", err)
	}

	var stale []staleKey
	for _, sa := range sas.Accounts {
		keys, err := svc.Projects.ServiceAccounts.Keys.List("projects/"+projectID+"/serviceAccounts/"+sa.Email).
			KeyTypes("USER_MANAGED").Context(ctx).Do()
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ! failed to list keys for %s: %v\n", sa.Email, err)
			continue
		}

		for _, k := range keys.Keys {
			created, perr := time.Parse(time.RFC3339, k.ValidAfterTime)
			if perr != nil {
				continue
			}
			if created.Before(cutoff) {
				stale = append(stale, staleKey{
					User:       sa.Email,
					KeyID:      k.Name,
					Status:     "USER_MANAGED",
					AgeDays:    int(time.Since(created).Hours() / 24),
					CreateDate: created,
				})
			}
		}
	}

	fmt.Printf("[gcp-scanner] found %d stale key(s) (threshold: %dd)\n", len(stale), thresholdDays)
	for _, s := range stale {
		fmt.Printf("  ✗ sa=%s key=%s age=%dd\n", s.User, s.KeyID, s.AgeDays)
	}

	if len(stale) > 0 && !dryRun {
		text := ""
		for _, s := range stale {
			text += fmt.Sprintf("• sa=%s key=%s age=%dd\n", s.User, s.KeyID, s.AgeDays)
		}
		if err := notify.Send("GCP service account stale keys detected", text); err != nil {
			return fmt.Errorf("notify: %w", err)
		}
	}

	return nil
}

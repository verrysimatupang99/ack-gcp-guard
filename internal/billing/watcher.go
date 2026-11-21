// Package billing implements the billing-watcher worker, which queries the
// GCP Billing BigQuery export and alerts when a day's spend exceeds the
// previous day by a configurable percentage.
package billing

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"

	"github.com/si_juling/ack-gcp-guard/pkg/notify"
)

type dailyTotal struct {
	Day   time.Time
	Total float64
}

// NewBillingWatcherCmd returns the billing-watcher worker command.
func NewBillingWatcherCmd() *cobra.Command {
	var spikePct float64
	var dryRun bool
	var lookbackDays int

	cmd := &cobra.Command{
		Use:   "billing-watcher",
		Short: "Alert when day-over-day billing spend exceeds a percentage",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := os.Getenv("GCP_PROJECT_ID")
			dataset := os.Getenv("GCP_BILLING_DATASET")
			table := os.Getenv("GCP_BILLING_TABLE")
			if projectID == "" || dataset == "" || table == "" {
				return fmt.Errorf("GCP_PROJECT_ID, GCP_BILLING_DATASET and GCP_BILLING_TABLE must be set")
			}

			client, err := bigquery.NewClient(cmd.Context(), projectID)
			if err != nil {
				return fmt.Errorf("bigquery client: %w", err)
			}
			defer client.Close()

			return watchBilling(cmd.Context(), client, dataset, table, spikePct, lookbackDays, dryRun)
		},
	}

	cmd.Flags().Float64Var(&spikePct, "spike-pct", 20, "alert when day-over-day increase exceeds this %")
	cmd.Flags().IntVar(&lookbackDays, "lookback-days", 7, "how many days to analyze")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print findings without alerting")

	return cmd
}

func watchBilling(ctx context.Context, client *bigquery.Client, dataset, table string, spikePct float64, days int, dryRun bool) error {
	totals, err := queryDailyTotals(ctx, client, dataset, table, days)
	if err != nil {
		return err
	}
	if len(totals) < 2 {
		fmt.Println("[billing-watcher] not enough data to compare days")
		return nil
	}

	for i := 1; i < len(totals); i++ {
		prev, cur := totals[i-1], totals[i]
		if prev.Total <= 0 {
			continue
		}

		pct := (cur.Total - prev.Total) / prev.Total * 100
		flag := ""
		if pct > spikePct {
			flag = " ⚠ ALERT"
			if !dryRun {
				msg := fmt.Sprintf("%s: $%.2f vs $%.2f (%.1f%%)",
					cur.Day.Format("2006-01-02"), cur.Total, prev.Total, pct)
				if err := notify.Send("Billing spike detected", msg); err != nil {
					return fmt.Errorf("notify: %w", err)
				}
			}
		}

		fmt.Printf("[billing-watcher] %s: $%.2f (prev $%.2f, %+.1f%%)%s\n",
			cur.Day.Format("2006-01-02"), cur.Total, prev.Total, pct, flag)
	}

	return nil
}

func queryDailyTotals(ctx context.Context, client *bigquery.Client, dataset, table string, days int) ([]dailyTotal, error) {
	q := client.Query(fmt.Sprintf(
		"SELECT DATE(usage_start_time) AS day, SUM(cost) AS total "+
			"FROM `%s.%s` "+
			"WHERE usage_start_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL %d DAY) "+
			"GROUP BY day ORDER BY day", dataset, table, days))

	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("query billing table: %w", err)
	}

	var out []dailyTotal
	for {
		var row struct {
			Day   bigquery.NullDate
			Total float64
		}
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read billing row: %w", err)
		}
		if row.Day.Valid {
			out = append(out, dailyTotal{Day: row.Day.Date.In(time.UTC), Total: row.Total})
		}
	}

	return out, nil
}

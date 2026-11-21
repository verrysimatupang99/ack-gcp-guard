package scanner

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	ram "github.com/alibabacloud-go/ram-20150501/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/spf13/cobra"

	"github.com/si_juling/ack-gcp-guard/pkg/notify"
)

const (
	maxThrottleRetries = 5
	defaultRegion       = "cn-hangzhou"
)

// staleKey is a cloud credential flagged by a scanner.
type staleKey struct {
	User       string
	KeyID      string
	Status     string
	AgeDays    int
	CreateDate time.Time
}

// NewRAMScannerCmd returns the ram-scanner worker command.
func NewRAMScannerCmd() *cobra.Command {
	var thresholdDays int
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "ram-scanner",
		Short: "Flag Alibaba RAM access keys older than N days",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newRAMClient()
			if err != nil {
				return err
			}
			return scanRAMUsers(cmd.Context(), client, thresholdDays, dryRun)
		},
	}

	cmd.Flags().IntVar(&thresholdDays, "threshold-days", 90, "flag keys older than this many days")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print findings without alerting")

	return cmd
}

func newRAMClient() (*ram.Client, error) {
	accessKey := os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID")
	accessSecret := os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET")
	if accessKey == "" || accessSecret == "" {
		return nil, fmt.Errorf("ALIBABA_CLOUD_ACCESS_KEY_ID and ALIBABA_CLOUD_ACCESS_KEY_SECRET must be set")
	}

	region := os.Getenv("ALIBABA_CLOUD_REGION")
	if region == "" {
		region = defaultRegion
	}

	config := &openapi.Config{
		AccessKeyId:     tea.String(accessKey),
		AccessKeySecret: tea.String(accessSecret),
	}
	config.Endpoint = tea.String(fmt.Sprintf("ram.%s.aliyuncs.com", region))

	return ram.NewClient(config)
}

func scanRAMUsers(ctx context.Context, client *ram.Client, thresholdDays int, dryRun bool) error {
	cutoff := time.Now().AddDate(0, 0, -thresholdDays)

	users, err := listAllUsers(ctx, client)
	if err != nil {
		return fmt.Errorf("list RAM users: %w", err)
	}

	var stale []staleKey
	for _, u := range users {
		userName := tea.StringValue(u.UserName)
		keys, err := listUserKeys(ctx, client, userName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ! failed to list keys for %s: %v\n", userName, err)
			continue
		}
		for _, k := range keys {
			created, perr := time.Parse("2006-01-02T15:04:05Z", tea.StringValue(k.CreateDate))
			if perr != nil {
				continue
			}
			if created.Before(cutoff) {
				stale = append(stale, staleKey{
					User:       userName,
					KeyID:      tea.StringValue(k.AccessKeyId),
					Status:     tea.StringValue(k.Status),
					AgeDays:    int(time.Since(created).Hours() / 24),
					CreateDate: created,
				})
			}
		}
	}

	fmt.Printf("[ram-scanner] found %d stale key(s) (threshold: %dd)\n", len(stale), thresholdDays)
	for _, s := range stale {
		fmt.Printf("  ✗ user=%s key=%s age=%dd status=%s\n", s.User, s.KeyID, s.AgeDays, s.Status)
	}

	if len(stale) > 0 && !dryRun {
		text := ""
		for _, s := range stale {
			text += fmt.Sprintf("• user=%s key=%s age=%dd status=%s\n", s.User, s.KeyID, s.AgeDays, s.Status)
		}
		if err := notify.Send("Alibaba RAM stale keys detected", text); err != nil {
			return fmt.Errorf("notify: %w", err)
		}
	}

	return nil
}

func listAllUsers(ctx context.Context, client *ram.Client) ([]*ram.ListUsersResponseBodyUsersUser, error) {
	var all []*ram.ListUsersResponseBodyUsersUser
	marker := ""
	runtime := &util.RuntimeOptions{}

	for {
		req := &ram.ListUsersRequest{}
		if marker != "" {
			req.Marker = tea.String(marker)
		}

		resp, err := withRetry(func() (*ram.ListUsersResponse, error) {
			return client.ListUsersWithOptions(req, runtime)
		})
		if err != nil {
			return nil, err
		}

		all = append(all, resp.Body.Users.User...)
		if resp.Body.IsTruncated == nil || !tea.BoolValue(resp.Body.IsTruncated) {
			break
		}
		marker = tea.StringValue(resp.Body.Marker)
	}

	return all, nil
}

func listUserKeys(ctx context.Context, client *ram.Client, userName string) ([]*ram.ListAccessKeysResponseBodyAccessKeysAccessKey, error) {
	req := &ram.ListAccessKeysRequest{UserName: tea.String(userName)}
	runtime := &util.RuntimeOptions{}

	resp, err := withRetry(func() (*ram.ListAccessKeysResponse, error) {
		return client.ListAccessKeysWithOptions(req, runtime)
	})
	if err != nil {
		return nil, err
	}

	return resp.Body.AccessKeys.AccessKey, nil
}

// withRetry retries throttled Alibaba OpenAPI calls with exponential backoff.
func withRetry[T any](fn func() (T, error)) (T, error) {
	var zero T
	var lastErr error

	for attempt := 0; attempt < maxThrottleRetries; attempt++ {
		res, err := fn()
		if err == nil {
			return res, nil
		}
		lastErr = err
		if !isThrottle(err) {
			return zero, err
		}

		backoff := time.Duration(1<<attempt) * time.Second
		fmt.Fprintf(os.Stderr, "  ! throttled, retrying in %s (attempt %d/%d)\n", backoff, attempt+1, maxThrottleRetries)
		time.Sleep(backoff)
	}

	return zero, fmt.Errorf("throttling persisted after %d retries: %w", maxThrottleRetries, lastErr)
}

// isThrottle reports whether an Alibaba SDK error is a throttling error.
func isThrottle(err error) bool {
	if sdkErr, ok := err.(*tea.SDKError); ok && sdkErr.Code != nil {
		code := strings.ToLower(tea.StringValue(sdkErr.Code))
		return strings.Contains(code, "throttl")
	}
	return false
}

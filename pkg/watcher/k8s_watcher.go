// Package watcher implements the k8s-watcher worker, which watches pods in
// ACK/GKE clusters and alerts when a container is stuck in CrashLoopBackOff.
package watcher

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	watchtools "k8s.io/client-go/tools/watch"

	"github.com/si_juling/ack-gcp-guard/pkg/notify"
)

const crashLoopReason = "CrashLoopBackOff"

// NewK8sWatcherCmd returns the k8s-watcher worker command.
func NewK8sWatcherCmd() *cobra.Command {
	var namespace string
	var allNamespaces bool
	var kubeconfig string

	cmd := &cobra.Command{
		Use:   "k8s-watcher",
		Short: "Watch CrashLoopBackOff pods in ACK/GKE and alert",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8sClient(kubeconfig)
			if err != nil {
				return err
			}
			return watchCrashLoop(cmd.Context(), client, namespace, allNamespaces)
		},
	}

	cmd.Flags().StringVar(&namespace, "namespace", "", "namespace to watch (default: all)")
	cmd.Flags().BoolVar(&allNamespaces, "all-namespaces", false, "watch all namespaces")
	cmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig (default: $KUBECONFIG or ~/.kube/config)")

	return cmd
}

func k8sClient(kubeconfig string) (*kubernetes.Clientset, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}
	configOverrides := &clientcmd.ConfigOverrides{}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("create kubernetes client: %w", err)
	}
	return client, nil
}

func watchCrashLoop(ctx context.Context, client *kubernetes.Clientset, namespace string, allNamespaces bool) error {
	ns := namespace
	if allNamespaces || ns == "" {
		ns = metav1.NamespaceAll
	}

	// Field selector: only watch pods that are running (a pod in
	// CrashLoopBackOff is still in phase Running).
	fieldSelector := fields.OneTermEqualSelector("status.phase", "Running").String()
	listOptions := metav1.ListOptions{FieldSelector: fieldSelector}

	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return client.CoreV1().Pods(ns).List(ctx, listOptions)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return client.CoreV1().Pods(ns).Watch(ctx, listOptions)
		},
	}

	w, err := watchtools.NewRetryWatcher("", lw)
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}

	fmt.Printf("[k8s-watcher] watching pods in namespace %q for CrashLoopBackOff...\n", ns)

	// Stop on SIGINT/SIGTERM.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		w.Stop()
	}()

	for event := range w.ResultChan() {
		pod, ok := event.Object.(*corev1.Pod)
		if !ok {
			continue
		}
		if container := crashLoopContainer(pod); container != "" {
			fmt.Printf("  ⚠ CrashLoopBackOff: %s/%s (container %s)\n", pod.Namespace, pod.Name, container)
			_ = notify.Send("CrashLoopBackOff detected",
				fmt.Sprintf("Pod %s/%s is in CrashLoopBackOff (container %s)", pod.Namespace, pod.Name, container))
		}
	}

	return nil
}

// crashLoopContainer returns the name of the first container stuck in
// CrashLoopBackOff, or "" if none.
func crashLoopContainer(pod *corev1.Pod) string {
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Waiting != nil && strings.EqualFold(status.State.Waiting.Reason, crashLoopReason) {
			return status.Name
		}
	}
	return ""
}

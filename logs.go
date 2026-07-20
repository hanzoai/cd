// logs.go — GET /v1/deploy/{name}/logs: the app's current pod logs, streamed from
// the newest running pod via the typed CoreV1 GetLogs subresource. The operator
// labels the workload it renders for an App CR with
// app.kubernetes.io/instance=<name>, so that selects the app's pods; the
// most-recently-started pod is read (the current rollout). Optional ?container=
// selects a container; ?tail= bounds the lines. Never fabricates output — an
// unreachable cluster or absent pod yields an honest 200 with an empty tail + the
// reason, not invented logs.
package deploy

import (
	"context"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hanzoai/cloud"
	"github.com/zap-proto/zip"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	logTailDefault  = int64(400)
	logTailMax      = int64(5000)
	logMaxBytes     = 512 << 10
	logReadDeadline = 8 * time.Second
)

// appLogs streams the newest app pod's logs. It resolves the app's namespace, then
// selects pods by the operator's instance label. 200 always (with content or an
// honest empty note) so a dashboard poll never errors on a not-yet-scheduled pod.
func appLogs(s *cloud.Service[state], c *zip.Ctx) error {
	if err := ready(s); err != nil {
		return err
	}
	name := reqName(c)
	if !appNameRE.MatchString(name) {
		return zip.ErrBadRequest("name must be a DNS-1123 label")
	}
	ns, err := resolveNamespace(s, c, name)
	if err != nil {
		return err
	}
	tail := logTailDefault
	if q := strings.TrimSpace(c.Query("tail")); q != "" {
		if n, e := strconv.ParseInt(q, 10, 64); e == nil && n > 0 {
			if n > logTailMax {
				n = logTailMax
			}
			tail = n
		}
	}
	container := strings.TrimSpace(c.Query("container"))

	logs, pod, ok := podLogs(s, c.Context(), ns, "app.kubernetes.io/instance="+name, container, tail)
	res := map[string]any{"application": ns + "/" + name, "pod": pod, "container": container, "logs": logs}
	if !ok {
		res["note"] = "no running pod logs available yet (pod not scheduled, or the cluster/pod is unreachable)"
	}
	return c.JSON(http.StatusOK, res)
}

// podLogs finds the newest pod matching selector in ns and streams its logs
// (tail-bounded, byte-capped, time-boxed). Returns (logs, podName, ok). ok=false on
// no typed client / no pod / read failure — the caller states the honest fallback.
func podLogs(s *cloud.Service[state], ctx context.Context, ns, selector, container string, tail int64) (string, string, bool) {
	if s.State.clientset == nil {
		return "", "", false
	}
	rctx, cancel := context.WithTimeout(ctx, logReadDeadline)
	defer cancel()

	pods, err := s.State.clientset.CoreV1().Pods(ns).List(rctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil || len(pods.Items) == 0 {
		return "", "", false
	}
	pod := newestPod(pods.Items)
	if pod == "" {
		return "", "", false
	}
	opts := &corev1.PodLogOptions{TailLines: &tail}
	if container != "" {
		opts.Container = container
	}
	stream, err := s.State.clientset.CoreV1().Pods(ns).GetLogs(pod, opts).Stream(rctx)
	if err != nil {
		return "", pod, false
	}
	defer stream.Close()

	data, err := io.ReadAll(io.LimitReader(stream, int64(logMaxBytes)+1))
	if err != nil && len(data) == 0 {
		return "", pod, false
	}
	out := string(data)
	if len(out) > logMaxBytes {
		out = out[len(out)-logMaxBytes:]
		if nl := strings.IndexByte(out, '\n'); nl >= 0 && nl < len(out)-1 {
			out = out[nl+1:]
		}
		out = "… (truncated to the most recent " + strconv.Itoa(logMaxBytes>>10) + " KiB)\n" + out
	}
	if strings.TrimSpace(out) == "" {
		return "", pod, false
	}
	return out, pod, true
}

// newestPod returns the name of the most-recently-started pod (by startTime, then
// creationTimestamp; ties break on name for stability).
func newestPod(pods []corev1.Pod) string {
	sort.Slice(pods, func(i, j int) bool {
		ti, tj := podTime(pods[i]), podTime(pods[j])
		if ti.Equal(tj) {
			return pods[i].Name > pods[j].Name
		}
		return ti.After(tj)
	})
	return pods[0].Name
}

func podTime(p corev1.Pod) time.Time {
	if p.Status.StartTime != nil {
		return p.Status.StartTime.Time
	}
	return p.CreationTimestamp.Time
}

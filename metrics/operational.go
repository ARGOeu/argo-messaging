package metrics

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ARGOeu/argo-messaging/stores"
	"github.com/ARGOeu/argo-messaging/tracectx"
	log "github.com/sirupsen/logrus"
)

func GetUsageCPUMem(ctx context.Context, store stores.Store) (MetricList, error) {
	pid := os.Getpid()
	pidstr := strconv.FormatInt(int64(pid), 10)
	out, err := exec.Command("ps", "-p", pidstr, "-o", "%cpu").Output()
	if err != nil {
		log.WithFields(
			log.Fields{
				"trace_id": tracectx.FromContext(ctx),
				"type":     "service_log",
			},
		).Error(err.Error())
	}

	// Take cli output and split it by new line chars
	cpuOut := strings.Split(string(out[:]), "\n")
	log.WithFields(
		log.Fields{
			"trace_id": tracectx.FromContext(ctx),
			"type":     "service_log",
		},
	).Info("CPU extracted value:", cpuOut[1])
	cpuVal, err := strconv.ParseFloat(strings.TrimSpace(cpuOut[1]), 64)
	if err != nil {
		log.WithFields(
			log.Fields{
				"trace_id": tracectx.FromContext(ctx),
				"type":     "service_log",
			},
		).Error(err.Error())
	}

	out2, err := exec.Command("ps", "-p", pidstr, "-o", "%mem").Output()
	if err != nil {
		log.WithFields(
			log.Fields{
				"trace_id": tracectx.FromContext(ctx),
				"type":     "service_log",
			},
		).Error(err.Error())
	}

	// Take cli output and split it by new line chars
	memOut := strings.Split(string(out2[:]), "\n")
	log.WithFields(
		log.Fields{
			"trace_id": tracectx.FromContext(ctx),
			"type":     "service_log",
		},
	).Info("MEM extracted value:", memOut[1])
	memVal, err := strconv.ParseFloat(strings.TrimSpace(memOut[1]), 64)
	if err != nil {
		log.WithFields(
			log.Fields{
				"trace_id": tracectx.FromContext(ctx),
				"type":     "service_log",
			},
		).Error(err.Error())
	}

	host, err := os.Hostname()
	if err != nil {
		log.WithFields(
			log.Fields{
				"trace_id": tracectx.FromContext(ctx),
				"type":     "service_log",
			},
		).Error(err.Error())
	}

	store.InsertOpMetric(ctx, host, cpuVal, memVal)
	result := store.GetOpMetrics(ctx)
	ml := MetricList{Metrics: []Metric{}}
	for _, v := range result {
		m := NewOpNodeCPU(v.Hostname, v.CPU, GetTimeNowZulu())
		m2 := NewOpNodeMEM(v.Hostname, v.MEM, GetTimeNowZulu())
		ml.Metrics = append(ml.Metrics, m)
		ml.Metrics = append(ml.Metrics, m2)
	}

	return ml, err
}

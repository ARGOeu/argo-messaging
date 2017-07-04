package metrics

import (
	"os"
	"os/exec"
	"strconv"

	"github.com/ARGOeu/argo-messaging/stores"
	log "github.com/Sirupsen/logrus"
)

func GetUsageCPU(store stores.Store) (MetricList, error) {
	pid := os.Getpid()
	pidstr := strconv.FormatInt(int64(pid), 10)
	out, err := exec.Command("ps", "-p", pidstr, "-o", "%cpu").Output()
	if err != nil {
		log.Error(err)
	}

	cpuVal, err := strconv.ParseFloat(string(out[:len(out)]), 64)
	if err != nil {
		log.Error(err)
	}

	host, err := os.Hostname()
	if err != nil {
		log.Error(err)
	}

	store.InsertOpMetric(host, cpuVal, 0.0)
	result := store.GetOpMetrics()
	ml := MetricList{Metrics: []Metric{}}
	for _, v := range result {
		m := NewOpNodeCPU(v.Hostname, v.CPU, GetTimeNowZulu())
		ml.Metrics = append(ml.Metrics, m)
	}

	return ml, err
}

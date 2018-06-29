package metrics

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ARGOeu/argo-messaging/stores"
	log "github.com/sirupsen/logrus"
)

func GetUsageCpuMem(store stores.Store) (MetricList, error) {
	pid := os.Getpid()
	pidstr := strconv.FormatInt(int64(pid), 10)
	out, err := exec.Command("ps", "-p", pidstr, "-o", "%cpu").Output()
	if err != nil {
		log.Error(err)

	}

	// Take cli output and split it by new line chars
	cpuOut := strings.Split(string(out[:len(out)]), "\n")
	log.Info("CPU extracted value:", cpuOut[1])
	cpuVal, err := strconv.ParseFloat(strings.TrimSpace(cpuOut[1]), 64)
	if err != nil {
		log.Error(err)
	}

	out2, err := exec.Command("ps", "-p", pidstr, "-o", "%mem").Output()
	if err != nil {
		log.Error(err)
	}

	// Take cli output and split it by new line chars
	memOut := strings.Split(string(out2[:len(out2)]), "\n")
	log.Info("MEM extracted value:", memOut[1])
	memVal, err := strconv.ParseFloat(strings.TrimSpace(memOut[1]), 64)
	if err != nil {
		log.Error(err)
	}

	host, err := os.Hostname()
	if err != nil {
		log.Error(err)
	}

	store.InsertOpMetric(host, cpuVal, memVal)
	result := store.GetOpMetrics()
	ml := MetricList{Metrics: []Metric{}}
	for _, v := range result {
		m := NewOpNodeCPU(v.Hostname, v.CPU, GetTimeNowZulu())
		m2 := NewOpNodeMEM(v.Hostname, v.MEM, GetTimeNowZulu())
		ml.Metrics = append(ml.Metrics, m)
		ml.Metrics = append(ml.Metrics, m2)
	}

	return ml, err
}

package k8sCustom

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

var sessionRowLimitRead int = 5

// TODO func save session data points create trend data ⬆ ⬇
type GetUtilizationStdOut struct {
	namespace    string
	NamespaceCPU int64
	NamespaceMem float64
	Workloads    []Wl
	Nodes        []Nd
	Kubecnf      string
	Clusterep    string
	Tick         time.Duration
	SessionDir   string
}

type Param struct {
	Tmpdir        string
	Action        string
	Debug         int
	Namespace     string
	LabelSelector string
	Sortby        string
	Indexes       map[string]Indexfile
}

type Wl struct {
	namespace   string
	Workload    string
	Workloadcpu int64
	Workloadmem float64
}

type Nd struct {
	Node          string
	Nodecpu       int64
	Nodemem       float64
	Nodecpupct    float64
	Nodemempct    float64
	Nodecpucalloc int64
	Nodememalloc  float64
}

// func debounceRun(fn func()) {
// 	if debounceTimer != nil {
// 		debounceTimer.Stop()
// 	}

// 	debounceTimer = time.AfterFunc(300*time.Millisecond, fn)
// }

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true // file exists
	}
	if os.IsNotExist(err) {
		return false // file does not exist
	}
	return false // some other error (e.g., permission)
}

type Indexfile struct {
	DirPath          string
	Filename         string
	fullFilenamePath string
}

type IIndexes interface {
	write(content string)
}

func (i *Indexfile) dirCreate() *Indexfile {
	err := os.MkdirAll(i.DirPath, 0755)
	if err != nil {
		panic(err)
	}

	return i
}

func (i *Indexfile) fileCreate() (*os.File, error) {
	filenamePath := fmt.Sprintf("%s/%s", i.DirPath, i.Filename)
	file, err := os.Create(filenamePath)
	if err != nil {
		panic(err)
	}
	i.fullFilenamePath = filenamePath
	return file, nil
}

func (i *Indexfile) SessionInit() {
	i.dirCreate().fileCreate()
}

func (i *Indexfile) write(content string) {
	file, err := os.OpenFile(i.fullFilenamePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.WriteString(fmt.Sprintf("%s\n", content))
}

func (i *Indexfile) parseIndexAvg(groupBy string, colIndex int) float64 {
	file, err := os.Open(i.fullFilenamePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	var window [][]string
	for _, row := range records {
		if len(row) == 0 {
			continue
		}
		if row[0] != groupBy {
			continue
		}

		window = append(window, row)
		if len(window) > sessionRowLimitRead {
			window = window[1:]
		}
	}

	var sum float64
	for _, row := range window {
		if len(row) <= colIndex {
			continue
		}
		val, err := strconv.ParseFloat(row[colIndex], 64)
		if err != nil {
			continue
		}
		sum += val
	}

	if len(window) == 0 {
		return 0
	}

	return sum / float64(len(window))
}

// TODO identify if 0 is from actual average | 0 if absence of data
func (ins *Indexfile) getTrendIndicator(groupColKey string, cpuUsage int64, memUsage float64, cpuDiffLimit float64, memDiffLimit float64) (cpuTrend string, memTrend string) {
	cpuLastAvg := ins.parseIndexAvg(groupColKey, 1)
	memLastAvg := ins.parseIndexAvg(groupColKey, 2)

	if cpuLastAvg <= 0 || memUsage <= 0 {
		return "", ""
	}

	cpuTrendns_ := cpuUsage - int64(cpuLastAvg)
	memTrendns_ := memUsage - memLastAvg
	cpuTrendns := ""
	if math.Abs(float64(cpuTrendns_)) > cpuDiffLimit { // capture 500milicores difference
		if cpuTrendns_ < 0 {
			cpuTrendns = "\033[31m⬇\033[0m"
		} else {
			cpuTrendns = "\033[32m⬆\033[0m"
			// cpuTrendns = fmt.Sprintf("\033[32m⬆\033[0m | %d - %d | %f <> %f", cpuUsage, int64(cpuLastAvg), math.Abs(float64(cpuTrendns_)), cpuDiffLimit)
		}
	}
	memTrendns := ""
	if math.Abs(float64(memTrendns_)) > memDiffLimit { // capture 500Mi difference
		if memTrendns_ < 0 {
			memTrendns = "\033[31m⬇\033[0m"
		} else {
			memTrendns = "\033[32m⬆\033[0m"
		}
	}

	return cpuTrendns, memTrendns
}

func ConfigInit(kubeconfig string) (*kubernetes.Clientset, *metricsclient.Clientset, string) {
	config, err := rest.InClusterConfig() // gets auto mounted secrets via automountServiceAccountToken: false

	if err != nil {
		if !fileExists(kubeconfig) {
			if home := homedir.HomeDir(); home != "" {
				kubeconfig = filepath.Join(home, ".kube", "config")
			}
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatalf("Error building kubeconfig: %v", err)
		}
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
	}

	// Create clientset for metrics
	metricsClient, err := metricsclient.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating metricsclientset: %v", err)
	}

	return clientset, metricsClient, kubeconfig
}

func (gtu *GetUtilizationStdOut) DisplayData(dsType string, param Param) {
	switch strings.ToLower(dsType) {
	case "watchutil":
		clearScreen()
		now := time.Now()

		// * CONFIG *
		tableCnf := tablewriter.NewWriter(os.Stdout)
		tableCnf.Header([]string{"Configs"})
		tableCnf.Append([]string{fmt.Sprintf("Date/Time: %s \nConfig: %s \nCluster: %s \nSort by: %s\nRefresh Rate: %0.fs\nSession file:%s", now.Format("2006-01-02 15:04:05"), gtu.Kubecnf, gtu.Clusterep, param.Sortby, gtu.Tick.Seconds(), gtu.SessionDir)})
		tableCnf.Render()
		// * CONFIG *

		// * NAMESPACE *
		ins := param.Indexes["namespace"]
		cpuTrendns, memTrendns := ins.getTrendIndicator("ALL", gtu.NamespaceCPU, gtu.NamespaceMem, 500, 0.5)

		tableNs := tablewriter.NewWriter(os.Stdout)
		tableNs.Header([]string{"namespace", "CPU", "MEM"})
		ns := "ALL"
		if gtu.namespace != "" {
			ns = gtu.namespace
		}
		tableNs.Append([]string{ns, fmt.Sprintf("%dm %s", gtu.NamespaceCPU, cpuTrendns), fmt.Sprintf("%1.fGi %s", gtu.NamespaceMem, memTrendns)})
		tableNs.Render()
		ins.write(fmt.Sprintf("%s,%d,%f", ns, gtu.NamespaceCPU, gtu.NamespaceMem))
		// * NAMESPACE *

		// * WORKLOADS *
		tableWorkloads := tablewriter.NewWriter(os.Stdout)
		tableWorkloads.Header([]string{"namespace", "Workload", "CPU", "MEM"})
		iworkloads := param.Indexes["workloads"]
		for _, w := range gtu.Workloads {
			cpuTrend, memTrend := ins.getTrendIndicator(w.Workload, w.Workloadcpu, w.Workloadmem, 500, 500)

			tableWorkloads.Append([]string{w.namespace, w.Workload, fmt.Sprintf("%dm %s", w.Workloadcpu, cpuTrend), fmt.Sprintf("%1.fMi %s", w.Workloadmem, memTrend)})
			iworkloads.write(fmt.Sprintf("%s,%d,%f", w.Workload, w.Workloadcpu, w.Workloadmem))
		}
		tableWorkloads.Render()
		// * WORKLOADS *

		// * NODEs *
		tableNodes := tablewriter.NewWriter(os.Stdout)
		tableNodes.Header([]string{"Node", "CPU", "MEM", "CPU(%)", "MEM(%)", "CPU Capacity", "MEM Capacity"})
		inodes := param.Indexes["nodes"]
		for _, n := range gtu.Nodes {
			cpuTrend, memTrend := ins.getTrendIndicator(n.Node, int64(math.Ceil(n.Nodecpupct)), n.Nodemempct, 5, 5)

			tableNodes.Append([]string{n.Node, fmt.Sprintf("%dm", n.Nodecpu), fmt.Sprintf("%1.fGi", n.Nodemem), fmt.Sprintf("%.1f %s", n.Nodecpupct, cpuTrend), fmt.Sprintf("%.1f %s", n.Nodemempct, memTrend), fmt.Sprintf("%dm", n.Nodecpucalloc), fmt.Sprintf("%1.fGi", n.Nodememalloc)})
			inodes.write(fmt.Sprintf("%s,%f,%f", n.Node, n.Nodecpupct, n.Nodemempct))
		}
		tableNodes.Render()
		// * NODEs *

		fmt.Printf("Ctrl + C to exit...")
	default:
		log.Fatalf("Display Type not recognized\n")
	}
}

func (gtu *GetUtilizationStdOut) GetUtilization(clientset *kubernetes.Clientset, metricsClient *metricsclient.Clientset, param Param) {
	gtu.Clusterep = clientset.RESTClient().Get().URL().Host
	switch strings.ToLower(param.LabelSelector) {
	case "":
		gtu.GetUtilizationNamespace(metricsClient, param.Namespace)
		gtu.GetUtilizationWorkload(clientset, metricsClient, param.Namespace, "")
	default:
		gtu.GetUtilizationWorkload(clientset, metricsClient, param.Namespace, param.LabelSelector)
	}
}

func (gtu *GetUtilizationStdOut) GetUtilizationNamespace(metricsClient *metricsclient.Clientset, ns string) {
	podMetricsList, err := metricsClient.MetricsV1beta1().PodMetricses(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	var totalCPU int64
	var totalMem int64

	for _, pod := range podMetricsList.Items {
		for _, c := range pod.Containers {
			cpuQty := c.Usage.Cpu().MilliValue() // millicores
			memQty := c.Usage.Memory().Value()   // bytes

			totalCPU += cpuQty
			totalMem += memQty
		}
	}

	gtu.namespace = ns
	gtu.NamespaceCPU = totalCPU
	gtu.NamespaceMem = float64(totalMem / 1024 / 1024 / 1024)
}

func (gtu *GetUtilizationStdOut) GetUtilizationWorkload(clientset *kubernetes.Clientset, metricsClient *metricsclient.Clientset, ns string, l string) {
	pods, err := clientset.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{LabelSelector: l})
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	metrics, err := metricsClient.MetricsV1beta1().PodMetricses(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	metricsMap := make(map[string]map[string]int64)
	for _, m := range metrics.Items {
		var cpu, mem int64
		for _, c := range m.Containers {
			cpu += c.Usage.Cpu().MilliValue()
			mem += c.Usage.Memory().Value()
		}
		metricsMap[m.Name] = map[string]int64{
			"cpu": cpu,
			"mem": mem,
		}
	}

	type wlAttib struct {
		namespace string
		Cpu       int64
		Mem       int64
	}

	workloads := make(map[string]wlAttib)
	for _, pod := range pods.Items {
		var totalCPU, totalMem int64
		if val, ok := metricsMap[pod.Name]; ok {
			totalCPU = val["cpu"]
			totalMem = val["mem"]
		}

		if len(pod.OwnerReferences) == 0 { // handle no owner pods // kubectl Debug node/k8s-worker-oci --image=busybox:stable -n kube-system -- sleep 3600
			key := "(no-owner)"
			if _, ok := workloads[key]; !ok {
				w := workloads[key]
				w.Cpu += totalCPU
				w.Mem += totalMem
				w.namespace = pod.Namespace
				workloads[key] = w
			}
			w := workloads[key]
			w.Cpu += totalCPU
			w.Mem += totalMem
			w.namespace = pod.Namespace
			workloads[key] = w
			continue
		}

		for i := 0; i < len(pod.OwnerReferences); i++ {
			owner := fmt.Sprintf("%s/%s", pod.OwnerReferences[i].Kind, pod.OwnerReferences[i].Name)
			if _, ok := workloads[owner]; !ok {
				w := workloads[owner] // copy of struct
				w.Cpu += totalCPU
				w.Mem += totalMem
				w.namespace = pod.Namespace
				workloads[owner] = w // reassign value to original struct
			}
			w := workloads[owner]
			w.Cpu += totalCPU
			w.Mem += totalMem
			w.namespace = pod.Namespace
			workloads[owner] = w
		}
	}

	gtu.Workloads = []Wl{}
	for k, v := range workloads {
		workload := Wl{
			namespace:   v.namespace,
			Workload:    k,
			Workloadcpu: v.Cpu,
			Workloadmem: float64(v.Mem / 1024 / 1024),
		}

		gtu.Workloads = append(gtu.Workloads, workload)
	}

	gtu.GetUtilizationNodes(clientset, metricsClient)
}

func (gtu *GetUtilizationStdOut) GetUtilizationNodes(clientset *kubernetes.Clientset, metricsClient *metricsclient.Clientset) {
	nodeSpecMap := make(map[string]map[string]int64)
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	for _, n := range nodes.Items {
		cpu := n.Status.Allocatable.Cpu().MilliValue()
		mem := n.Status.Allocatable.Memory().Value()

		nodeSpecMap[n.Name] = map[string]int64{
			"cpu": cpu,
			"mem": mem,
		}
	}

	metrics, err := metricsClient.MetricsV1beta1().NodeMetricses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	metricsMap := make(map[string]map[string]int64)
	for _, m := range metrics.Items {
		var cpu, mem int64
		cpu += m.Usage.Cpu().MilliValue()
		mem += m.Usage.Memory().Value()

		metricsMap[m.Name] = map[string]int64{
			"cpu": cpu,
			"mem": mem,
		}
	}

	var cpuPct, memPct float64
	var nodesStats []Nd
	for i, v := range metricsMap {
		cpuUsage := v["cpu"]
		memUsage := float64(v["mem"] / 1024 / 1024 / 1024)

		cpuAlloc := nodeSpecMap[i]["cpu"]
		memAlloc := float64(nodeSpecMap[i]["mem"] / 1024 / 1024 / 1024)

		cpuPct = float64(cpuUsage) / float64(cpuAlloc) * 100
		memPct = float64(memUsage) / float64(memAlloc) * 100

		nodeState := Nd{
			Node:          i,
			Nodecpu:       cpuUsage,
			Nodemem:       memUsage,
			Nodecpupct:    cpuPct,
			Nodemempct:    memPct,
			Nodecpucalloc: cpuAlloc,
			Nodememalloc:  memAlloc,
		}
		nodesStats = append(nodesStats, nodeState)
	}

	gtu.Nodes = nodesStats
}

func (gtu *GetUtilizationStdOut) Sort(param Param) {
	sort.Slice(gtu.Nodes, func(i, j int) bool {
		if param.Sortby == "mem" {
			return gtu.Nodes[i].Nodemem > gtu.Nodes[j].Nodemem
		} else {
			return gtu.Nodes[i].Nodecpu > gtu.Nodes[j].Nodecpu
		}
	})

	sort.Slice(gtu.Workloads, func(i, j int) bool {
		if param.Sortby == "mem" {
			return gtu.Workloads[i].Workloadmem > gtu.Workloads[j].Workloadmem
		} else {
			return gtu.Workloads[i].Workloadcpu > gtu.Workloads[j].Workloadcpu
		}
	})
}

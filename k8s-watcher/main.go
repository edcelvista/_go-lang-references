package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

var debug *int
var action *string
var namespace *string
var labelSelector *string

func parseFlags() {
	debug = flag.Int("debug", 1, "Minimal logs (default) = 1 | Basic info (normal output) = 2 | More details about the request/response = 3-4 | Debug level (headers, request path, timing) = 5-6 | Verbose debugging, request/response bodies = 7-8 | Very verbose, internal client-go details = 9 | Maximum verbosity, internal details of client-go and API calls")
	action = flag.String("action", "", "watchPods, watchEvents")
	namespace = flag.String("namespace", "", "target namespace")
	labelSelector = flag.String("label", "", "label of target workload object")

	flag.Parse()
}

func main() {
	parseFlags()
	clientset, metricsClientset := configInit()

	log.Printf("⚡️ Action: %s | 📦 %s\n", strings.ToLower(*action), *namespace)
	switch strings.ToLower(*action) {
	case "watchpods":
		watchPods(clientset, *namespace)
	case "watchevents":
		watchEvents(clientset, *namespace)
	case "getutil":
		getUtilization(clientset, metricsClientset, *namespace, *labelSelector)
	default:
		log.Fatalf("Action not recognized %s", *action)
	}
}

func configInit() (*kubernetes.Clientset, *metricsclient.Clientset) {
	config, err := rest.InClusterConfig() // gets auto mounted secrets via automountServiceAccountToken: false

	if err != nil {
		// Load kubeconfig (for out-of-cluster use)
		var kubeconfig *string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()

		// Build config
		log.Printf("📁 KubeConfig: %v \n", string(*kubeconfig))
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig) // fall into kubeconfig file
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

	return clientset, metricsClient
}

func watchPods(clientset *kubernetes.Clientset, namespace string) {
	ctx := context.Background()

	// Create a watcher
	watcher, err := clientset.CoreV1().Pods(namespace).Watch(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error creating pod watcher: %v", err)
	}
	defer watcher.Stop()

	log.Println("🔭 Watching for pod changes...")

	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Added:
			pod := event.Object.(*v1.Pod)
			log.Printf("🟢 Pod Added: %s/%s Node: %s\n", pod.Namespace, pod.Name, pod.Spec.NodeName)

		case watch.Modified:
			pod := event.Object.(*v1.Pod)

			containers := "[ "
			for i := 0; i < len(pod.Status.ContainerStatuses); i++ {
				containerName := pod.Status.ContainerStatuses[i].Name
				isReady := pod.Status.ContainerStatuses[i].Ready
				state := pod.Status.ContainerStatuses[i].State

				var status string

				switch {
				case state.Running != nil:
					status = "Running"
				case state.Waiting != nil:
					status = state.Waiting.Reason
				case state.Terminated != nil:
					status = fmt.Sprintf("%s %d",
						state.Terminated.Reason,
						state.Terminated.ExitCode,
					)
				default:
					status = "Unknown"
				}

				// last entry
				separator := ""
				if i == len(pod.Status.ContainerStatuses)-1 {
					separator = ""
				} else {
					separator = ","
				}

				containers += fmt.Sprintf("%s: %s %s%s ", containerName, strconv.FormatBool(isReady), status, separator)
			}
			containers += "]"

			log.Printf("🟡 Pod Modified: %s/%s | Phase: %s | Containers: %s\n", pod.Namespace, pod.Name, pod.Status.Phase, containers)

		case watch.Deleted:
			pod := event.Object.(*v1.Pod)
			log.Printf("🔴 Pod Deleted: %s/%s Node: %s\n", pod.Namespace, pod.Name, pod.Spec.NodeName)

		case watch.Error:
			log.Println("⚠️ Watch error received, reconnecting...")
			time.Sleep(2 * time.Second)
			watchPods(clientset, namespace) // restart watch
			return
		}
	}
}

func watchEvents(clientset *kubernetes.Clientset, namespace string) {
	ctx := context.Background()

	// Create a watcher
	watcher, err := clientset.CoreV1().Events(namespace).Watch(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error creating pod watcher: %v", err)
	}
	defer watcher.Stop()

	log.Println("🔭 Watching for events changes...")

	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Added:
			event := event.Object.(*v1.Event)
			log.Printf("🟢 Event Added: %s/%s %s\n", event.Namespace, event.Name, event.Message)

		case watch.Modified:
			event := event.Object.(*v1.Event)
			log.Printf("🟡 Event Modified: %s/%s (Reason: %s) \n", event.Namespace, event.Name, event.Reason)

		case watch.Deleted:
			event := event.Object.(*v1.Event)
			log.Printf("🔴 Event Deleted: %s/%s\n", event.Namespace, event.Name)

		case watch.Error:
			log.Println("⚠️ Watch error received, reconnecting...")
			time.Sleep(2 * time.Second)
			watchPods(clientset, namespace) // restart watch
			return
		}
	}
}

func getUtilization(clientset *kubernetes.Clientset, metricsClient *metricsclient.Clientset, namespace string, labelSelector string) {
	switch strings.ToLower(labelSelector) {
	case "":
		getUtilizationNamespace(metricsClient, namespace)
		getUtilizationWorkload(clientset, metricsClient, namespace, "")
	default:
		getUtilizationWorkload(clientset, metricsClient, namespace, labelSelector)
	}
}

func getUtilizationNamespace(metricsClient *metricsclient.Clientset, namespace string) {
	podMetricsList, err := metricsClient.MetricsV1beta1().
		PodMetricses(namespace).
		List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		panic(err)
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

	log.Printf("📦 Namespace: %s\n", namespace)
	log.Printf("🤖 Namespace CPU: %dm\n", totalCPU)
	log.Printf("🤖 Namespace Memory: %dMi\n", totalMem/1024/1024)
}

func getUtilizationWorkload(clientset *kubernetes.Clientset, metricsClient *metricsclient.Clientset, namespace string, labelSelector string) {
	pods, err := clientset.CoreV1().
		Pods(namespace).
		List(context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})

	if err != nil {
		panic(err)
	}

	metrics, err := metricsClient.MetricsV1beta1().
		PodMetricses(namespace).
		List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		panic(err)
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

	workloads := make(map[string]map[string]int64)
	for _, pod := range pods.Items {

		var totalCPU, totalMem int64
		if val, ok := metricsMap[pod.Name]; ok {
			totalCPU = val["cpu"]
			totalMem = val["mem"]
		}

		if len(pod.OwnerReferences) == 0 { // handle no owner pods // kubectl debug node/k8s-worker-oci --image=busybox:stable -n kube-system -- sleep 3600
			workloads["no-owner"] = make(map[string]int64)
			workloads["no-owner"]["cpu"] += totalCPU
			workloads["no-owner"]["mem"] += totalMem
			continue
		}

		for i := 0; i < len(pod.OwnerReferences); i++ {
			owner := fmt.Sprintf("%s/%s", pod.OwnerReferences[i].Kind, pod.OwnerReferences[i].Name)
			if _, ok := workloads[owner]; !ok {
				workloads[owner] = make(map[string]int64)
			}
			workloads[owner]["cpu"] += totalCPU
			workloads[owner]["mem"] += totalMem
		}
	}

	for k, v := range workloads {
		log.Printf("🖥️ Workload: %s \n", k)
		log.Printf("🤖 CPU: %dm", v["cpu"])
		log.Printf("🤖 Mem: %dMi", v["mem"]/1024/1024)
	}
}

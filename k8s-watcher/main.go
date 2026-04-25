package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/eiannone/keyboard"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"

	k8sCustom "edcelvista.com/k8s/watcher/k8sCustom"
)

var (
	debug         int
	action        string
	refreshrate   int
	sortby        string
	namespace     string
	labelSelector string
	kubeconfig    string
	gtuOut        k8sCustom.GetUtilizationStdOut
	lastSort      string = "cpu"
	tick          time.Duration
)

// func echoToJson(st interface{}) {
// 	stJson, err := json.Marshal(st)
// 	if err != nil {
// 		fmt.Printf("Error: %s \n", err)
// 	}
// 	fmt.Printf("%s\n", stJson)
// }

func parseArgs() {
	flag.IntVar(&debug, "debug", 1, "Minimal logs (default) = 1 | Basic info (normal output) = 2 | More details about the request/response = 3-4 | Debug level (headers, request path, timing) = 5-6 | Verbose debugging, request/response bodies = 7-8 | Very verbose, internal client-go details = 9 | Maximum verbosity, internal details of client-go and API calls")
	flag.StringVar(&action, "action", "", "watchPods, watchEvents, watchutil")
	flag.IntVar(&refreshrate, "refreshrate", 10, "5, 10")
	flag.StringVar(&sortby, "sortby", "cpu", "cpu, mem")
	flag.StringVar(&namespace, "namespace", "", "target namespace")
	flag.StringVar(&labelSelector, "label", "", "label of target workload object")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "kubeconfig file")
	flag.Parse()
}

func tmpDir() string {
	tmpDir, err := os.MkdirTemp("", "cnf-*")
	if err != nil {
		log.Fatalf("Error creating temp folder: %v", err)
	}
	return tmpDir
}

func main() {
	parseArgs()
	tmpDir := tmpDir()

	defer os.RemoveAll(tmpDir)
	param := k8sCustom.Param{
		Tmpdir:        tmpDir,
		Debug:         debug,
		Action:        action,
		Namespace:     namespace,
		LabelSelector: labelSelector,
		Sortby:        sortby,
	}

	clientset, metricsClientset, kubecnf := k8sCustom.ConfigInit(kubeconfig)
	gtuOut.Kubecnf = kubecnf

	switch strings.ToLower(action) {
	case "watchpods":
		watchPods(clientset, namespace)
	case "watchevents":
		watchEvents(clientset, namespace)
	case "watchutil":
		fmt.Printf("Preparing Indexes...\n")
		param.Indexes = map[string]k8sCustom.Indexfile{}

		nsIndex := k8sCustom.Indexfile{
			DirPath:  fmt.Sprintf("%s", tmpDir),
			Filename: "namespaces",
			Debug:    debug,
		}
		nsIndex.SessionInit()
		param.Indexes["namespace"] = nsIndex

		worloadsIndex := k8sCustom.Indexfile{
			DirPath:  fmt.Sprintf("%s", tmpDir),
			Filename: "workloads",
			Debug:    debug,
		}
		worloadsIndex.SessionInit()
		param.Indexes["workloads"] = worloadsIndex

		nodesIndex := k8sCustom.Indexfile{
			DirPath:  fmt.Sprintf("%s", tmpDir),
			Filename: "nodes",
			Debug:    debug,
		}
		nodesIndex.SessionInit()
		param.Indexes["nodes"] = nodesIndex

		initRun := make(chan int32, 1)

		fmt.Printf("Triggering Init Run...\n")
		initRun <- 1

		tick = time.Duration(refreshrate)
		duration := tick * time.Second
		ticker := time.NewTicker(duration)
		defer ticker.Stop()

		resize := make(chan os.Signal, 1)
		signal.Notify(resize, syscall.SIGWINCH)

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

		sort := make(chan string, 1)
		go func() {
			if err := keyboard.Open(); err != nil {
				log.Fatal(err)
			}
			defer keyboard.Close()

			for {
				char, key, err := keyboard.GetKey()
				if err != nil {
					log.Fatal(err)
				}

				if key == keyboard.KeyCtrlC {
					quit <- os.Interrupt
					return
				}

				if key == keyboard.KeySpace {
					resize <- syscall.SIGWINCH
				}

				switch char {
				case 'm', 'M':
					sort <- "mem"

				case 'c', 'C':
					sort <- "cpu"
				}
			}
		}()

		gtuOut.Tick = duration
		gtuOut.SessionDir = tmpDir

		for {
			select {
			case <-initRun:
				watchUtilization(clientset, metricsClientset, param, true, lastSort)
			case <-ticker.C:
				watchUtilization(clientset, metricsClientset, param, true, lastSort)
			case <-resize:
				watchUtilization(clientset, metricsClientset, param, false, lastSort)
			case sortVal := <-sort: // store in variable read it once
				lastSort = sortVal
				watchUtilization(clientset, metricsClientset, param, false, lastSort)
			case <-quit:
				os.RemoveAll(tmpDir)
				fmt.Printf("\nStopped\n")
				return
			}
		}
	default:
		flag.Usage()
	}
}

func watchPods(clientset *kubernetes.Clientset, namespace string) {
	ctx := context.Background()

	watcher, err := clientset.CoreV1().Pods(namespace).Watch(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error creating pod watcher: %v", err)
	}
	defer watcher.Stop()

	log.Printf("🔭 Watching for pod changes...\n")
	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Added:
			pod := event.Object.(*v1.Pod)
			fmt.Printf("🟢 Pod Added: %s/%s Node: %s\n", pod.Namespace, pod.Name, pod.Spec.NodeName)

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

			fmt.Printf("🟡 Pod Modified: %s/%s | Phase: %s | Containers: %s\n", pod.Namespace, pod.Name, pod.Status.Phase, containers)

		case watch.Deleted:
			pod := event.Object.(*v1.Pod)
			fmt.Printf("🔴 Pod Deleted: %s/%s Node: %s\n", pod.Namespace, pod.Name, pod.Spec.NodeName)

		case watch.Error:
			log.Printf("⚠️ Watch error received, reconnecting...\n")
			time.Sleep(2 * time.Second)
			watchPods(clientset, namespace) // restart watch
			return
		}
	}
}

func watchEvents(clientset *kubernetes.Clientset, namespace string) {
	ctx := context.Background()

	watcher, err := clientset.CoreV1().Events(namespace).Watch(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error creating pod watcher: %v", err)
	}
	defer watcher.Stop()

	log.Printf("🔭 Watching for events changes...")
	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Added:
			event := event.Object.(*v1.Event)
			fmt.Printf("🟢 Event Added: %s/%s %s\n", event.Namespace, event.Name, event.Message)

		case watch.Modified:
			event := event.Object.(*v1.Event)
			fmt.Printf("🟡 Event Modified: %s/%s (Reason: %s) \n", event.Namespace, event.Name, event.Reason)

		case watch.Deleted:
			event := event.Object.(*v1.Event)
			fmt.Printf("🔴 Event Deleted: %s/%s\n", event.Namespace, event.Name)

		case watch.Error:
			log.Printf("⚠️ Watch error received, reconnecting...\n")
			time.Sleep(2 * time.Second)
			watchPods(clientset, namespace) // restart watch
			return
		}
	}
}

func watchUtilization(clientset *kubernetes.Clientset, metricsClientset *metricsclient.Clientset, param k8sCustom.Param, fetch bool, sort string) {
	if fetch {
		gtuOut.GetUtilization(clientset, metricsClientset, param)
	}
	param.Sortby = sort
	gtuOut.Sort(param)
	gtuOut.DisplayData("watchutil", param)
}

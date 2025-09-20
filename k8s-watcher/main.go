package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	config, err := rest.InClusterConfig()
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
		fmt.Printf("KubeConfig: %v \n", string(*kubeconfig))
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			log.Fatalf("Error building kubeconfig: %v", err)
		}
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
	}

	// Start watching pods in all namespaces
	watchPods(clientset)
}

func watchPods(clientset *kubernetes.Clientset) {
	ctx := context.Background()

	// Create a watcher
	watcher, err := clientset.CoreV1().Pods("").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error creating pod watcher: %v", err)
	}
	defer watcher.Stop()

	log.Println("🔭 Watching for pod changes...")

	for event := range watcher.ResultChan() {
		switch event.Type {
		case watch.Added:
			pod := event.Object.(*v1.Pod)
			fmt.Printf("🟢 Pod Added: %s/%s\n", pod.Namespace, pod.Name)

		case watch.Modified:
			pod := event.Object.(*v1.Pod)
			fmt.Printf("🟡 Pod Modified: %s/%s (Phase: %s)\n", pod.Namespace, pod.Name, pod.Status.Phase)

		case watch.Deleted:
			pod := event.Object.(*v1.Pod)
			fmt.Printf("🔴 Pod Deleted: %s/%s\n", pod.Namespace, pod.Name)

		case watch.Error:
			log.Println("⚠️ Watch error received, reconnecting...")
			time.Sleep(2 * time.Second)
			watchPods(clientset) // restart watch
			return
		}
	}
}

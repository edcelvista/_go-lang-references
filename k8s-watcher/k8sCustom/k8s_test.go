package k8sCustom

import (
	"testing"
	"time"
)

var gtuOut GetUtilizationStdOut

func TestReceiver(t *testing.T) {
	gtuOut := GetUtilizationStdOut{
		namespace:    "test",
		NamespaceCPU: 1,
		NamespaceMem: 1.0,
		Kubecnf:      "/tmp/test/namespace",
		Clusterep:    "https://local.test",
		Tick:         time.Duration.Abs(10),
		SessionDir:   "/tmp/test",
	}

	gtuOut.Workloads = append(gtuOut.Workloads, Wl{
		namespace:   "test",
		Workload:    "test/test",
		Workloadcpu: 1,
		Workloadmem: 1.0,
	})

	gtuOut.Nodes = append(gtuOut.Nodes, Nd{
		Node:          "test",
		Nodecpu:       1,
		Nodemem:       1.0,
		Nodecpupct:    1.0,
		Nodemempct:    1.0,
		Nodecpucalloc: 1,
		Nodememalloc:  1.0,
	})

	param := Param{
		Tmpdir:        "/tmp/test",
		Action:        "watchutil",
		Debug:         4,
		Namespace:     "test",
		LabelSelector: "",
		Sortby:        "cpu",
	}

	param.Indexes = map[string]Indexfile{}

	nsIndex := Indexfile{
		DirPath:  "/tmp/test",
		Filename: "namespaces",
	}
	nsIndex.SessionInit()
	param.Indexes["namespace"] = nsIndex

	worloadsIndex := Indexfile{
		DirPath:  "/tmp/test",
		Filename: "workloads",
	}
	worloadsIndex.SessionInit()
	param.Indexes["workloads"] = worloadsIndex

	nodesIndex := Indexfile{
		DirPath:  "/tmp/test",
		Filename: "nodes",
	}
	nodesIndex.SessionInit()
	param.Indexes["nodes"] = nodesIndex

	gtuOut.DisplayData("watchutil", param)
}

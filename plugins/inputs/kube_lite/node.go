package kube_lite

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ericchiang/k8s/apis/core/v1"

	"github.com/influxdata/telegraf"
)

var (
	nodeMeasurement = "kube_node"
)

func collectNodes(ctx context.Context, acc telegraf.Accumulator, ks *KubernetesState) {
	list, err := ks.client.getNodes(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, n := range list.Items {
		if err = ks.gatherNode(*n, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ks *KubernetesState) gatherNode(n v1.Node, acc telegraf.Accumulator) error {
	fields := map[string]interface{}{}
	tags := map[string]string{
		"name": *n.Metadata.Name,
	}

	for resourceName, val := range n.Status.Capacity {
		switch resourceName {
		// todo: cpu or cpu_cores
		case "cpu":
			// todo: better way to get value
			fields["status_capacity_cpu_cores"] = atoi(*val.String_)
		case "memory":
			// todo: better way to get value, verify
			fields["status_capacity_"+sanitizeLabelName(resourceName)+"_bytes"] = atoi(*val.String_)
		case "pods":
			// todo: better way to get value
			fields["status_capacity_pods"] = atoi(*val.String_)
		}
	}

	for resourceName, val := range n.Status.Allocatable {
		switch resourceName {
		case "cpu":
			fields["status_allocatable_cpu_cores"] = atoi(*val.String_)
		case "memory":
			fields["status_allocatable_"+sanitizeLabelName(string(resourceName))+"_bytes"] = atoi(*val.String_)
		case "pods":
			fields["status_allocatable_pods"] = atoi(*val.String_)
		}
	}

	acc.AddFields(nodeMeasurement, fields, tags)
	return nil
}

func atoi(s string) int64 {
	i, err := strconv.Atoi(s)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return int64(i)
}
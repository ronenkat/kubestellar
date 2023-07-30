/*
Copyright 2022 The KubeStellar Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	klog "k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/mcclient"
)

func main() {
	ctx := context.Background()
	logger := klog.Background()
	ctx = klog.NewContext(ctx, logger)

	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"

	managementClusterConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		logger.Error(err, "failed to build management cluster config")
		panic(err)
	}

	// Create the MC aware client --> initiate the underlying MC aware library
	// The library actively watches for LC updates, and maintain an updated list of accessible clusters
	mcclient, err := mcclient.NewMultiCluster(ctx, managementClusterConfig)
	if err != nil {
		logger.Error(err, "get client failed")
		panic(err)
	}

	clusterName := "ks-lc4"

	// Demonstrate a Watch() on a logical cluster
	// Using the mcclient to get access to a LC directly (clientset, informer, etc..)
	watcher, err := mcclient.Cluster(clusterName).Kube().CoreV1().ConfigMaps(metav1.NamespaceDefault).Watch(ctx, metav1.ListOptions{})
	if err == nil {
		for {
			select {
			case <-ctx.Done():
				watcher.Stop()
			case event, ok := <-watcher.ResultChan():
				if !ok {
					watcher.Stop()
					return
				}
				if event.Type == "ADDED" {
					cm := event.Object.(*corev1.ConfigMap)
					logger.Info("New configmap detected", "name", cm.Name, "cluster", clusterName)
				}
			}
		}
	}
}

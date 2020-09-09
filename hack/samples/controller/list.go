// Copyright (c) 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/vmware-tanzu/net-operator-api/api/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"

	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var testEnv *envtest.Environment

const namespace string = "default"

// List VirtualMachines in a target cluster to stdout using a controller client
func main() {
	fmt.Printf("Starting test env...\n")
	testClient, err := startTestEnv()
	if err != nil {
		panic(err)
	}
	defer func() {
		fmt.Printf("Stopping test env...\n")
		testEnv.Stop()
	}()

	netOpClient, err := getNetOpClient(testClient)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Populating test env...\n")
	err = populateTestEnv(netOpClient, "test-if1")
	if err != nil {
		panic(err)
	}
	err = populateTestEnv(netOpClient, "test-if2")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Listing NetworkInterfaces:\n")
	netIfList := v1alpha1.NetworkInterfaceList{}
	err = netOpClient.List(context.TODO(), &netIfList)
	if err != nil {
		panic(err)
	}
	for _, netIf := range netIfList.Items {
		fmt.Printf("- %s\n", netIf.GetName())
	}
}

// Get a net-operator-api client from the generated clientset
func getNetOpClient(config *rest.Config) (ctrlClient.Client, error) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)
	client, err := ctrlClient.New(config, ctrlClient.Options{
		Scheme: scheme,
	})
	return client, err
}

func startTestEnv() (*rest.Config, error) {
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "config", "crd", "bases"),
		},
	}

	return testEnv.Start()
}

func populateTestEnv(client ctrlClient.Client, name string) error {
	newNetIf := v1alpha1.NetworkInterface{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return client.Create(context.TODO(), &newNetIf)
}

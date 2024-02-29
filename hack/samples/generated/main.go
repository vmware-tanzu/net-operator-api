// Copyright (c) 2020-2024 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/vmware-tanzu/net-operator-api/api/v1alpha1"
	netopv1alpha1 "github.com/vmware-tanzu/net-operator-api/pkg/client/clientset_generated/clientset/typed/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var testEnv *envtest.Environment

const namespace string = "default"

// List VirtualMachines in a target cluster to stdout using the generated client
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
	netIfList, err := netOpClient.NetworkInterfaces(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, netIf := range netIfList.Items {
		fmt.Printf("- %s\n", netIf.GetName())
	}
}

// Get a net-operator-api client from the generated clientset
func getNetOpClient(client *rest.Config) (*netopv1alpha1.NetoperatorV1alpha1Client, error) {
	return netopv1alpha1.NewForConfig(client)
}

func startTestEnv() (*rest.Config, error) {
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "config", "crd", "bases"),
		},
	}

	return testEnv.Start()
}

func populateTestEnv(client *netopv1alpha1.NetoperatorV1alpha1Client, name string) error {
	_, err := client.NetworkInterfaces(namespace).Create(context.TODO(),
		&v1alpha1.NetworkInterface{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		},
		metav1.CreateOptions{},
	)
	return err
}

// SPDX-License-Identifier: Apache-2.0
// Based on Code: https://github.com/johandry/klient

package client

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateNamespace creates a namespace with the given name
func (c *Client) CreateNamespace(namespace string) error {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"name": namespace,
			},
		},
	}
	_, err := c.Clientset.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	// if errors.IsAlreadyExists(err) {
	// 	// If it failed because the NS is already there, then do not return such error
	// 	return nil
	// }

	return err
}

// DeleteNamespace deletes the namespace with the given name
func (c *Client) DeleteNamespace(namespace string) error {
	return c.Clientset.CoreV1().Namespaces().Delete(context.TODO(), namespace, metav1.DeleteOptions{})
}

// NodesReady returns the number of nodes ready
func (c *Client) NodesReady() (ready int, total int, err error) {
	nodes, err := c.Clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return 0, 0, err
	}
	total = len(nodes.Items)
	if total == 0 {
		return 0, 0, nil
	}
	for _, n := range nodes.Items {
		for _, c := range n.Status.Conditions {
			if c.Type == "Ready" && c.Status == "True" {
				ready++
				break
			}
		}
	}

	return ready, len(nodes.Items), nil
}

// Version returns the cluster version. It can be used to verify if the cluster
// is reachable. It will return an error if failed to connect.
func (c *Client) Version() (string, error) {
	v, err := c.Clientset.ServerVersion()
	if err != nil {
		return "", err
	}

	return v.String(), nil
}

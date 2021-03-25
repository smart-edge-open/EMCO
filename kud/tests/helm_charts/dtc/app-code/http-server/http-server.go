// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	podIP, hostname := getPodDetails()
	fmt.Fprintf(w, "%s Hello from http-server with the pod IP - %s and podname - %s", t.Format("2006-01-02 15:04:05"), podIP, hostname)
}

func getPodDetails() (string, string) {

	ip := ""
	name := ""
	for {
		if ip != "" {
			break
		}

		addrs, err := net.InterfaceAddrs()
		if err != nil {
			fmt.Printf("unable to get the pod IP - %s\n", err.Error())
		}
		for _, address := range addrs {
			// check the address type and if it is not a loopback the display it
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ip = ipnet.IP.String()
					name = os.Getenv("HOSTNAME")
					return ip, name
				}
			}
		}
		time.Sleep(1 * time.Second)
	}

	return ip, name

}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":3333", nil)
}

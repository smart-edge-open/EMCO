// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func main() {
	httpHost := os.Getenv("SERVERDOMAIN")
	for {
		resp, err := http.Get(httpHost)
		if err != nil {
			fmt.Printf("unable to connect to http server - %s\n", err.Error())
		} else {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("unable to read the http get response- %s\n", err.Error())
			}
			fmt.Println("get:\n", string(body))
			resp.Body.Close()
		}
		time.Sleep(5 * time.Second)
	}

}

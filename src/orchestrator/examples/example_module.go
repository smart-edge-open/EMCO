// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package test

import (
	moduleLib "github.com/open-ness/EMCO/src/orchestrator/pkg/module"
	"log"
)

// ExampleClient_Project to test Project
func ExampleClient_Project() {
	// Get handle to the client
	c := moduleLib.NewClient()
	// Check if project is initialized
	if c.Project == nil {
		log.Println("Project is Uninitialized")
		return
	}
	// Perform operations on Project Module
	// POST request (exists == false)
	_, err := c.Project.CreateProject(moduleLib.Project{MetaData: moduleLib.ProjectMetaData{Name: "test", Description: "test", UserData1: "userData1", UserData2: "userData2"}}, false)
	if err != nil {
		log.Println(err)
		return
	}
	// PUT request (exists == true)
	_, err = c.Project.CreateProject(moduleLib.Project{MetaData: moduleLib.ProjectMetaData{Name: "test", Description: "test", UserData1: "userData1", UserData2: "userData2"}}, true)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = c.Project.GetProject("test")
	if err != nil {
		log.Println(err)
		return
	}
	err = c.Project.DeleteProject("test")
	if err != nil {
		log.Println(err)
	}
}

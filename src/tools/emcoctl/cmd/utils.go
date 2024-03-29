// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	neturl "net/url"
	"os"
	"strings"

	"text/template"

	"github.com/go-resty/resty/v2"
	"github.com/mitchellh/mapstructure"
	pkgerrors "github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var inputFiles []string
var valuesFiles []string
var token []string

type ResourceContext struct {
	Anchor string `json:"anchor" yaml:"anchor"`
}

type Metadata struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	UserData1   string `yaml:"userData1,omitempty" json:"userData1,omitempty"`
	UserData2   string `yaml:"userData2,omitempty" json:"userData2,omitempty"`
}

type emcoRes struct {
	Version string                 `yaml:"version" json:"version"`
	Context ResourceContext        `yaml:"resourceContext" json:"resourceContext"`
	Meta    Metadata               `yaml:"metadata" json:"metadata"`
	Spec    map[string]interface{} `yaml:"spec,omitempty" json:"spec,omitempty"`
	File    string                 `yaml:"file,omitempty" json:"file,omitempty"`
	Files   []string               `yaml:"files,omitempty" json:"files,omitempty"`
	Label   string                 `yaml:"label-name,omitempty" json:"label-name,omitempty"`
}

type emcoBody struct {
	Meta  Metadata               `json:"metadata,omitempty"`
	Label string                 `json:"label-name,omitempty"`
	Spec  map[string]interface{} `json:"spec,omitempty"`
}

type emcoCompositeAppSpec struct {
	Version string `json: "version"`
}

type Resources struct {
	anchor string
	body   []byte
	file   string
	files  []string
}

// RestyClient to use with CLI
type RestyClient struct {
	client *resty.Client
}

var Client RestyClient

// NewRestClient returns a rest client
func NewRestClient() RestyClient {
	// Create a Resty Client
	Client.client = resty.New()
	// Registering global Error object structure for JSON/XML request
	//Client.client.SetError(&Error{})
	return Client
}

// NewRestClientToken returns a rest client with token
func NewRestClientToken(token string) RestyClient {
	// Create a Resty Client
	Client.client = resty.New()
	// Bearer Auth Token for all request
	Client.client.SetAuthToken(token)
	// Registering global Error object structure for JSON/XML request
	//Client.client.SetError(&Error{})
	return Client
}

// readResources reads all the resources in the file provided
func readResources() []Resources {
	// TODO: Remove Assumption only one file
	// Open file and Parse to get all resources
	var resources []Resources
	var buf bytes.Buffer

	if len(valuesFiles) > 0 {
		//Apply template
		v, err := os.Open(valuesFiles[0])
		defer v.Close()
		if err != nil {
			fmt.Println("Error reading file", "error", err, "filename", valuesFiles[0])
			return []Resources{}
		}
		valDec := yaml.NewDecoder(v)
		var mapDoc interface{}
		if valDec.Decode(&mapDoc) != nil {
			fmt.Println("Values file format incorrect:", "error", err, "filename", valuesFiles[0])
			return []Resources{}
		}
		// Templatize
		t, err := template.ParseFiles(inputFiles[0])
		if err != nil {
			fmt.Println("Error reading file", "error", err, "filename", inputFiles[0])
			return []Resources{}
		}
		err = t.Execute(&buf, mapDoc)
		if err != nil {
			fmt.Println("execute: ", err)
			return []Resources{}
		}
	} else {
		f, err := os.Open(inputFiles[0])
		defer f.Close()
		if err != nil {
			fmt.Println("Error reading file", "error", err, "filename", inputFiles[0])
			return []Resources{}
		}
		io.Copy(&buf, f)
	}

	dec := yaml.NewDecoder(&buf)
	// Iterate through all resources in the file
	for {
		var doc emcoRes
		if err := dec.Decode(&doc); err != nil {
			if err.Error() != "EOF" {
				fmt.Println("Invalid input Yaml! Exiting..", err)
				// Exit executing
				os.Exit(1)
			}
			break
		}
		body := &emcoBody{Meta: doc.Meta, Spec: doc.Spec, Label: doc.Label}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			fmt.Println("Invalid input Yaml! Exiting..", err)
			// Exit executing
			os.Exit(1)
		}
		var res Resources
		if doc.File != "" {
			res = Resources{anchor: doc.Context.Anchor, body: jsonBody, file: doc.File}
		} else if len(doc.Files) > 0 {
			res = Resources{anchor: doc.Context.Anchor, body: jsonBody, files: doc.Files}
		} else {
			res = Resources{anchor: doc.Context.Anchor, body: jsonBody}
		}
		resources = append(resources, res)
	}
	return resources
}

//RestClientApply to post to server no multipart
func (r RestyClient) RestClientApply(anchor string, body []byte, put bool) error {
	var resp *resty.Response
	var err error
	var url string

	if put {
		if anchor, err = getUpdateUrl(anchor, body); err != nil {
			return err
		}
		if url, err = GetURL(anchor); err != nil {
			return err
		}
		// Put JSON string
		resp, err = r.client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(body).
			Put(url)
	} else {
		if url, err = GetURL(anchor); err != nil {
			return err
		}
		// Post JSON string
		resp, err = r.client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(body).
			Post(url)
	}
	if err != nil {
		fmt.Println(err)
		return err
	}
	if put {
		printOutput(url, "PUT", resp)
	} else {
		printOutput(url, "POST", resp)
	}
	if resp.StatusCode() >= 200 && resp.StatusCode() <= 299 {
		return nil
	}
	return pkgerrors.Errorf("Server Error")
}

//RestClientPut to post to server no multipart
func (r RestyClient) RestClientPut(anchor string, body []byte) error {
	if anchor == "" {
		return pkgerrors.Errorf("Anchor can't be empty")
	}
	s := strings.Split(anchor, "/")
	a := s[len(s)-1]
	if a == "instantiate" || a == "apply" || a == "approve" || a == "terminate" || a == "migrate" || a == "update" || a == "rollback" {
		// No put for these
		return nil
	}
	return r.RestClientApply(anchor, body, true)
}

//RestClientPost to post to server no multipart
func (r RestyClient) RestClientPost(anchor string, body []byte) error {
	return r.RestClientApply(anchor, body, false)
}

//RestClientMultipartApply to post to server with multipart
func (r RestyClient) RestClientMultipartApply(anchor string, body []byte, file string, put bool) error {
	var resp *resty.Response
	var err error
	var url string

	// Read file for multipart
	f, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("Error reading file", "error", err, "filename", file)
		return err
	}

	// Multipart Post
	formParams := neturl.Values{}
	formParams.Add("metadata", string(body))
	if put {
		if anchor, err = getUpdateUrl(anchor, body); err != nil {
			return err
		}
		if url, err = GetURL(anchor); err != nil {
			return err
		}
		resp, err = r.client.R().
			SetFileReader("file", "filename", bytes.NewReader(f)).
			SetFormDataFromValues(formParams).
			Put(url)
	} else {
		if url, err = GetURL(anchor); err != nil {
			return err
		}
		resp, err = r.client.R().
			SetFileReader("file", "filename", bytes.NewReader(f)).
			SetFormDataFromValues(formParams).
			Post(url)
	}
	if err != nil {
		fmt.Println(err)
		return err
	}
	if put {
		printOutput(url, "PUT", resp)
	} else {
		printOutput(url, "POST", resp)
	}
	if resp.StatusCode() >= 201 && resp.StatusCode() <= 299 {
		return nil
	}
	return pkgerrors.Errorf("Server Error")
}

//RestClientMultipartPut to post to server with multipart
func (r RestyClient) RestClientMultipartPut(anchor string, body []byte, file string) error {
	return r.RestClientMultipartApply(anchor, body, file, true)
}

//RestClientMultipartPost to post to server with multipart
func (r RestyClient) RestClientMultipartPost(anchor string, body []byte, file string) error {
	return r.RestClientMultipartApply(anchor, body, file, false)
}

func getFile(file string) ([]byte, string, error) {
	// Read file for multipart
	f, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("Error reading file", "error", err, "filename", file)
		return []byte{}, "", err
	}
	// Extract filename
	s := strings.TrimSuffix(file, "/")
	s1 := strings.Split(s, "/")
	name := s1[len(s1)-1]
	return f, name, nil
}

//RestClientMultipartApplyMultipleFiles to post to server with multipart
func (r RestyClient) RestClientMultipartApplyMultipleFiles(anchor string, body []byte, files []string, put bool) error {
	var f []byte
	var name string
	var err error
	var url string
	var resp *resty.Response

	req := r.client.R()
	// Multipart Post
	formParams := neturl.Values{}
	formParams.Add("metadata", string(body))
	// Add all files in the list
	for _, file := range files {
		f, name, err = getFile(file)
		if err != nil {
			return err
		}
		req = req.
			SetFileReader("files", name, bytes.NewReader(f))
	}
	if put {
		if anchor, err = getUpdateUrl(anchor, body); err != nil {
			return err
		}
		if url, err = GetURL(anchor); err != nil {
			return err
		}
		resp, err = req.
			SetFormDataFromValues(formParams).
			Put(url)
	} else {
		if url, err = GetURL(anchor); err != nil {
			return err
		}
		resp, err = req.
			SetFormDataFromValues(formParams).
			Post(url)
	}
	if err != nil {
		fmt.Println(err)
		return err
	}
	if put {
		printOutput(url, "PUT", resp)
	} else {
		printOutput(url, "POST", resp)
	}
	if resp.StatusCode() >= 200 && resp.StatusCode() <= 299 {
		return nil
	}
	return pkgerrors.Errorf("Server Error")
}

//RestClientMultipartPutMultipleFiles to post to server with multipart
func (r RestyClient) RestClientMultipartPutMultipleFiles(anchor string, body []byte, files []string) error {
	return r.RestClientMultipartApplyMultipleFiles(anchor, body, files, true)
}

//RestClientMultipartPostMultipleFiles to post to server with multipart
func (r RestyClient) RestClientMultipartPostMultipleFiles(anchor string, body []byte, files []string) error {
	return r.RestClientMultipartApplyMultipleFiles(anchor, body, files, false)
}

// RestClientGetAnchor returns get data from anchor
func (r RestyClient) RestClientGetAnchor(anchor string) error {
	url, err := GetURL(anchor)
	if err != nil {
		return err
	}
	s := strings.Split(anchor, "/")
	if len(s) >= 3 {
		a := s[len(s)-2]
		// Determine if multipart
		if a == "apps" || a == "profiles" || a == "clusters" || a == "resources" || a == "customizations" {
			// Supports only getting metadata
			resp, err := r.client.R().
				SetHeader("Accept", "application/json").
				Get(url)
			if err != nil {
				fmt.Println(err)
				return err
			}
			printOutput(url, "GET", resp)
			return nil
		}
	}
	resp, err := r.client.R().
		Get(url)
	if err != nil {
		fmt.Println(err)
		return err
	}
	printOutput(url, "GET", resp)
	return nil
}

func getUpdateUrl(anchor string, body []byte) (string, error) {
	var e emcoBody
	err := json.Unmarshal(body, &e)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	s := strings.Split(anchor, "/")
	a := s[len(s)-1]
	if e.Meta.Name != "" {
		name := e.Meta.Name
		anchor = anchor + "/" + name
		if a == "composite-apps" {
			var cav emcoCompositeAppSpec
			err := mapstructure.Decode(e.Spec, &cav)
			if err != nil {
				fmt.Println("Unable to decode CompositeApp Spec")
				return "", err
			}
			anchor = anchor + "/" + cav.Version
		}
	} else if e.Label != "" {
		anchor = anchor + "/" + e.Label
	}
	return anchor, nil
}

// RestClientGet gets resource
func (r RestyClient) RestClientGet(anchor string, body []byte) error {
	if anchor == "" {
		return pkgerrors.Errorf("Anchor can't be empty")
	}
	s := strings.Split(anchor, "/")
	a := s[len(s)-1]
	if a == "instantiate" || a == "apply" || a == "approve" || a == "terminate" || a == "migrate" || a == "update" || a == "rollback" {
		// No get for these
		return nil
	}
	c, err := getUpdateUrl(anchor, body)
	if err != nil {
		return err
	}
	return r.RestClientGetAnchor(c)
}

// RestClientDeleteAnchor returns all resource in the input file
func (r RestyClient) RestClientDeleteAnchor(anchor string) error {
	url, err := GetURL(anchor)
	if err != nil {
		return err
	}
	resp, err := r.client.R().Delete(url)
	if err != nil {
		fmt.Println(err)
		return err
	}
	printOutput(url, "DELETE", resp)
	return nil
}

// RestClientDelete calls rest delete command
func (r RestyClient) RestClientDelete(anchor string, body []byte) error {

	s := strings.Split(anchor, "/")
	a := s[len(s)-1]
	if a == "instantiate" {
		// Change instantiate to destroy
		s[len(s)-1] = "terminate"
		anchor = strings.Join(s[:], "/")
		return r.RestClientPost(anchor, []byte{})
	} else if a == "apply" {
		// Change apply to terminate
		s[len(s)-1] = "terminate"
		anchor = strings.Join(s[:], "/")
		return r.RestClientPost(anchor, []byte{})
	} else if a == "approve" || a == "status" || a == "migrate" || a == "update" || a == "rollback" {
		// No delete required for these
		return nil
	}
	var e emcoBody
	err := json.Unmarshal(body, &e)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if e.Meta.Name != "" {
		s := strings.Split(anchor, "/")
		a := s[len(s)-1]
		name := e.Meta.Name
		anchor = anchor + "/" + name
		if a == "composite-apps" {
			var cav emcoCompositeAppSpec
			err := mapstructure.Decode(e.Spec, &cav)
			if err != nil {
				fmt.Println("Unable to decode CompositeApp Spec")
				return err
			}
			anchor = anchor + "/" + cav.Version
		}
	} else if e.Label != "" {
		anchor = anchor + "/" + e.Label
	}
	return r.RestClientDeleteAnchor(anchor)
}

// GetURL reads the configuration file to get URL
func GetURL(anchor string) (string, error) {
	var baseUrl string
	s := strings.Split(anchor, "/")
	if len(s) < 1 {
		return "", fmt.Errorf("Invalid Anchor: %s", s)
	}

	switch s[0] {
	case "cluster-providers":
		if len(s) >= 5 && (s[4] == "networks" || s[4] == "provider-networks" ||
			s[4] == "apply" || s[4] == "terminate" || s[4] == "status") {
			baseUrl = GetNcmURL()
		} else {
			baseUrl = GetClmURL()
		}
	case "controllers":
		baseUrl = GetOrchestratorURL()
	case "clm-controllers":
		baseUrl = GetClmURL()
	case "dtc-controllers":
		baseUrl = GetDtcURL()
	case "projects":
		if len(s) >= 3 && s[2] == "logical-clouds" {
			baseUrl = GetDcmURL()
			break
		}
		// next two network-chains and sfc-clients must come before the
		// network-controller-intent
		if len(s) >= 10 && s[9] == "network-chains" {
			baseUrl = GetSfcURL()
			break
		}
		if len(s) >= 10 && s[9] == "sfc-clients" {
			baseUrl = GetSfcClientURL()
			break
		}
		if len(s) >= 8 && s[7] == "network-controller-intent" {
			baseUrl = GetOvnactionURL()
			break
		}
		if len(s) >= 8 && s[7] == "traffic-group-intents" {
			baseUrl = GetDtcURL()
			break
		}
		if len(s) >= 8 && s[7] == "generic-k8s-intents" {
			baseUrl = GetGacURL()
			break
		}
		if len(s) >= 8 && s[7] == "hpa-intents" {
			baseUrl = GetHpaPlacementURL()
			break
		}
		// All other paths go to Orchestrator
		baseUrl = GetOrchestratorURL()
	default:
		return "", fmt.Errorf("Invalid baseUrl: %s", baseUrl)
	}
	return (baseUrl + "/" + anchor), nil
}

func printOutput(url, op string, resp *resty.Response) {
	fmt.Println("---")
	fmt.Println(op, " --> URL:", url)
	fmt.Println("Response Code:", resp.StatusCode())
	if len(resp.Body()) > 0 {
		fmt.Println("Response:", resp)
	}
}

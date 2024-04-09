/*******************************************************************************
 * Copyright (c) 2024 Jan van Deventer
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v2.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-2.0/
 *
 * Contributors:
 *   Jan A. van Deventer, Luleå - initial implementation
 *   Thomas Hedeler, Hamburg - initial implementation
 ***************************************************************************SDG*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/vanDeventer/mbaigo/components"
	"github.com/vanDeventer/mbaigo/usecases"
)

// This is the main function for the OPC UA Client system
func main() {
	// prepare for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background()) // create a context that can be cancelled
	defer cancel()

	// instantiate the System
	sys := components.NewSystem("opcuac", ctx)

	// instantiate a template of the unit asset for configuration
	template := uaConfig{
		Name:          "PLC with OPC UA server",
		Details:       map[string][]string{"PLC": {"Prosys OPC UA Simulation Server"}, "Location": {"Line 1"}, "KKS":{"YLLCP001"}},
		ServerAddress: "opc.tcp://10.0.0.17:53530/OPCUA/SimulationServer",
		NodeList:      map[string][]string{"Node_Id": {}, "Browse_Name": {}, "Ref_Type": {}},
	}
	sys.UAsset = []components.UnitAsset{&template}

	// define the services that expose the capablities of the unit assets or nodes
	browse := components.Service{
		Definition:  "browse",
		SubPath:     "browse",
		Details:     map[string][]string{"Protocol": {"opc.tcp"}},
		RegPeriod:   61,
		Description: "provides the human readable (HTML) list (GET) of the nodes the OPC UA server holds, ",
	}

	access := components.Service{
		Definition:  "access",
		SubPath:     "access",
		Details:     map[string][]string{"Protocol": {"opc.tcp"}},
		RegPeriod:   30,
		Description: "accesses the OPC UA node to read (GET) the information or if posssible to write (PUT)[but not yet], ",
	}

	servs := components.Services{browse, access}

	// Instatiate the Capusle
	sys.Husk = &components.Husk{
		Description: "interacts with an OPC UA server",
		Details:     map[string][]string{"Developer": {"Arrowhead"}},
		ProtoPort:   map[string]int{"https": 0, "http": 9696, "coap": 0},
		DocLink:     "https://github.com/vanDeventer/mbaigo/tree/master/uaclient",
	}

	// Configure the syste5m
	rawResources, err := usecases.Configure(&sys, &servs)
	if err != nil {
		log.Fatalf("Configuration error: %v\n", err)
	}
	sys.UAsset = nil
	Resources := make(map[string]*UnitAsset)

	for _, raw := range rawResources {
		var cua uaConfig
		if err := json.Unmarshal(raw, &cua); err != nil {
			log.Fatalf("Resource configuration error: %+v\n", err)
		}
		nodelist, cleanup := newResource(&cua, &sys, &servs)
		defer cleanup() // This defers the cleanup (close OPC UA connection) to when main returns

		for _, node := range nodelist {
			sys.UAsset = append(sys.UAsset, &node)
			Resources[node.GetName()] = &node
		}
	}

	// Generate PKI keys and CSR to obtain a authentication certificate from the CA
	usecases.CreateCSR(&sys)

	// Register the (system) and its services
	usecases.RegisterServices(&sys)

	// start the http handler and server
	go usecases.HttpInterface(&sys)

	// wait for shutdown signal, and gracefully close properly goroutines with context
	<-sys.Sigs // wait for a SIGINT (Ctrl+C) signal
	fmt.Println("\nshuting down system", sys.Name)
	cancel()                    // cancel the context, signaling the goroutines to stop
	time.Sleep(3 * time.Second) // allow the go routines to be executed, which might take more time than the main routine to end
}

// Serving handles the resources services. NOTE: it exepcts those names from the request URL path
func (node *UnitAsset) Serving(w http.ResponseWriter, r *http.Request, servicePath string) {
	switch servicePath {
	case "browse":
		node.browse(w, r)
	case "access":
		node.access(w, r)

	default:
		http.Error(w, "Invalid service request [Do not modify the services subpath in the configurration file]", http.StatusBadRequest)
	}
}

func (node *UnitAsset) browse(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		node.browseNode(w)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)
	}
}

//

func (node *UnitAsset) access(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		vauleForm := node.read()
		usecases.HTTPProcessGetRequest(w, r, &vauleForm)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)
	}
}

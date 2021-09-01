// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	clmcontrollerpb "github.com/open-ness/EMCO/src/clm/pkg/grpc/controller-eventchannel"
	"github.com/open-ness/EMCO/src/hpa-plc/api"
	register "github.com/open-ness/EMCO/src/hpa-plc/pkg/grpc"
	clmControllerserver "github.com/open-ness/EMCO/src/hpa-plc/pkg/grpc/clmcontrollereventchannelserver"
	placementcontrollerserver "github.com/open-ness/EMCO/src/hpa-plc/pkg/grpc/hpaplacementcontrollerserver"
	plsctrlclientpb "github.com/open-ness/EMCO/src/orchestrator/pkg/grpc/placementcontroller"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/auth"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/config"
	contextDb "github.com/open-ness/EMCO/src/orchestrator/pkg/infra/contextdb"
	"github.com/open-ness/EMCO/src/orchestrator/pkg/infra/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
)

func startGrpcServer() error {
	tls := strings.Contains(config.GetConfiguration().GrpcEnableTLS, "enable")
	certFile := config.GetConfiguration().GrpcServerCert
	keyFile := config.GetConfiguration().GrpcServerKey

	_, port := register.GetServerHostPort()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Could not listen to port: [%v]", err)
	}
	log.Printf("Listening on grpc_port:[%v]\n", port)

	var opts []grpc.ServerOption
	if tls {
		if certFile == "" {
			certFile = testdata.Path("server.pem")
		}
		if keyFile == "" {
			keyFile = testdata.Path("server.key")
		}
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			log.Fatalf("Could not generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	grpcServer := grpc.NewServer(opts...)
	plsctrlclientpb.RegisterPlacementControllerServer(grpcServer, placementcontrollerserver.NewHpaPlacementControllerServer())
	clmcontrollerpb.RegisterClmControllerEventChannelServer(grpcServer, clmControllerserver.NewControllerEventchannelServer())

	log.Printf("Starting HPA PlacementController gRPC Server @ [%s]", lis.Addr().String())
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("hpaplacement grpc server is not serving %v", err)
	}
	return err
}

func main() {
	log.Printf("\nHPA PlacementController config @ [%v]\n", config.GetConfiguration())
	rand.Seed(time.Now().UnixNano())

	err := db.InitializeDatabaseConnection("mco")
	if err != nil {
		log.Println("Unable to initialize database connection...")
		log.Println(err)
		log.Fatalln("Exiting...")
	}
	err = contextDb.InitializeContextDatabase()
	if err != nil {
		log.Println("Unable to initialize database connection...")
		log.Println(err)
		log.Fatalln("Exiting...")
	}

	httpRouter := api.NewRouter(nil)
	loggedRouter := handlers.LoggingHandler(os.Stdout, httpRouter)
	httpServer := &http.Server{
		Handler: loggedRouter,
		Addr:    ":" + config.GetConfiguration().ServicePort,
	}
	log.Printf("\nStarting HPA PlacementController Http Server @ [%v]\n", httpServer.Addr)

	go func() {
		err := startGrpcServer()
		if err != nil {
			log.Fatalf("GRPC server failed to start")
		}
	}()

	connectionsClose := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		err := httpServer.Shutdown(context.Background())
		if err != nil {
			log.Fatalf("http server failed to shutdown")
		}
		close(connectionsClose)
	}()

	tlsConfig, err := auth.GetTLSConfig("ca.cert", "server.cert", "server.key")
	if err != nil {
		log.Println("Error Getting TLS Configuration. Starting without TLS...")
		log.Fatal(httpServer.ListenAndServe())
	} else {
		httpServer.TLSConfig = tlsConfig
		// empty strings because tlsconfig already has this information
		err = httpServer.ListenAndServeTLS("", "")
		if err != nil {
			log.Fatalf("http server Listening failed")
		}
	}
}

// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
	"k8s-terminal/handler"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	insecurePort        int
	insecureBindAddress net.IP
)

func parseFlags() {
	pflag.IntVar(&insecurePort, "insecure-port", 9090, "The port to listen to for incoming HTTP requests.")
	pflag.IPVar(&insecureBindAddress, "insecure-bind-address", net.IPv4(127, 0, 0, 1), "The IP address on which to serve the --port (set to 0.0.0.0 for all interfaces).")
	pflag.Parse()
}

func main() {
	parseFlags()

	mux := http.NewServeMux()

	mux.Handle("/api/sockjs/", sockjs.NewHandler("/api/sockjs", sockjs.DefaultOptions, handler.HandleTerminalSession))

	// http 服务
	httpServer := &http.Server{
		Addr: fmt.Sprintf("%s:%d", insecureBindAddress, insecurePort),
		Handler: mux,
	}

	// Shutdown gracefully on SIGTERM or SIGINT
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)

	idleConnsClosed := make(chan struct{})

	// 接收退出信号
	go func() {
		<-sigint
		// We received an interrupt signal, shut down.
		if err := httpServer.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			logrus.Infof("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("Serving insecurely on HTTP port: %d", insecurePort)

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		logrus.WithError(err).Fatal("http server start error")
	}

	<-idleConnsClosed
}

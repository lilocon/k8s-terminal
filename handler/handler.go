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

package handler

import (
	"encoding/json"
	"fmt"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

const END_OF_TRANSMISSION = "\u0004"

var clientManager = &KubernetesClientManager{
	clients: make(map[string]*kubernetes.Clientset),
}

func HandleTerminalSession(session sockjs.Session) {
	fmt.Print(session.ID())
	var (
		buf              string
		err              error
		handshakeMessage HandshakeMessage
		terminalSession  TerminalSession
	)

	if buf, err = session.Recv(); err != nil {
		//log.Printf("handleTerminalSession: can't Recv: %v", err)
		session.Close(1, fmt.Sprintf("can't Recv: %v", err))
		return
	}

	if err = json.Unmarshal([]byte(buf), &handshakeMessage); err != nil {
		//log.Printf("handleTerminalSession: can't UnMarshal (%v): %s", err, buf)
		session.Close(1, fmt.Sprintf("can't UnMarshal (%v): %s", err, buf))
		return
	}

	if handshakeMessage.Op != "bind" {
		//log.Printf("handleTerminalSession: expected 'bind' message, got: %s", buf)
		session.Close(1, fmt.Sprintf("expected 'bind' message, got: %s", buf))
		return
	}

	terminalSession = TerminalSession{
		sizeChan:      make(chan remotecommand.TerminalSize),
		sockJSSession: session,
	}

	fmt.Print(handshakeMessage)

	err = startProcess(terminalSession, handshakeMessage)

	if err != nil {
		session.Close(2, err.Error())
		return
	}

	session.Close(1, "Process exited")
}

// startProcess is called by handleAttach
// Executed cmd in the container specified in request and connects it up with the ptyHandler (a session)
func startProcess(ptyHandler PtyHandler, handshakeMessage HandshakeMessage) error {
	k8sClient, err := clientManager.getClient(handshakeMessage.Cluster)
	if err != nil {
		return err
	}
	cfg, err := clientManager.getRestClientConfig(handshakeMessage.Cluster)
	if err != nil {
		return err
	}

	req := k8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(handshakeMessage.Pod).
		Namespace(handshakeMessage.Namespace).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
		Container: handshakeMessage.Container,
		Command:   []string{"/bin/sh"},
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:             ptyHandler,
		Stdout:            ptyHandler,
		Stderr:            ptyHandler,
		TerminalSizeQueue: ptyHandler,
		Tty:               true,
	})
	if err != nil {
		return err
	}

	return nil
}

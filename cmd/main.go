// Copyright 2023 Milton Jesus aka miltlima

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"kubelearn/pkg/k8s"
	"kubelearn/pkg/resources"
	"kubelearn/pkg/utils"

	"github.com/gorilla/websocket"
	"k8s.io/client-go/dynamic"
)

// listenAddr is loopback-only: the /terminal endpoint runs shell commands, so
// the server must never be reachable from other machines.
const listenAddr = "127.0.0.1:8083"

// allowedOrigins holds the browser origins allowed to call the API and open
// the terminal WebSocket. Override the default (the local frontend) with a
// comma-separated list in KUBELEARN_ALLOWED_ORIGINS.
var allowedOrigins = parseAllowedOrigins(os.Getenv("KUBELEARN_ALLOWED_ORIGINS"))

func parseAllowedOrigins(raw string) map[string]bool {
	origins := map[string]bool{}
	for _, origin := range strings.Split(raw, ",") {
		origin = strings.TrimRight(strings.TrimSpace(origin), "/")
		if origin != "" {
			origins[origin] = true
		}
	}
	if len(origins) == 0 {
		origins["http://localhost:3000"] = true
		origins["http://127.0.0.1:3000"] = true
	}
	return origins
}

// originAllowed reports whether the request comes from an allowed origin.
// Browsers always send Origin on WebSocket handshakes, so a missing header is
// rejected too.
func originAllowed(r *http.Request) bool {
	return allowedOrigins[r.Header.Get("Origin")]
}

var upgrader = websocket.Upgrader{
	CheckOrigin: originAllowed,
}

// setupTargets are the Makefile targets run by /setup, in order: provision
// the cluster and switch the kubectl context, apply the lab manifests, then
// install the Gateway API used by the networking questions.
var setupTargets = []string{"infra-up", "manifests-apply", "gateway-api-install"}

// setupRunning guards against overlapping /setup runs.
var setupRunning atomic.Bool

// runMakeCommand runs a Makefile target in dir and logs the output.
// It is a variable so tests can replace it.
var runMakeCommand = func(dir, target string) error {
	cmd := exec.Command("make", target)
	cmd.Dir = dir
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd.Run()
}

// projectRoot returns the directory holding the project Makefile: the
// KUBELEARN_ROOT environment variable, or the working directory.
func projectRoot() (string, error) {
	root := os.Getenv("KUBELEARN_ROOT")
	if root == "" {
		root = "."
	}
	for _, name := range []string{"Makefile", "makefile"} {
		if info, err := os.Stat(filepath.Join(root, name)); err == nil && !info.IsDir() {
			return root, nil
		}
	}
	return "", fmt.Errorf("no Makefile found in %q: start the backend from the repository root or set KUBELEARN_ROOT", root)
}

// setupEnvironment provisions the lab in the background using the Makefile.
// It only accepts POST, since it changes infrastructure.
func setupEnvironment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	root, err := projectRoot()
	if err != nil {
		log.Printf("Cannot start environment setup: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !setupRunning.CompareAndSwap(false, true) {
		http.Error(w, "environment setup already running", http.StatusConflict)
		return
	}

	go func() {
		defer setupRunning.Store(false)
		log.Println("Starting environment setup...")
		for _, target := range setupTargets {
			if err := runMakeCommand(root, target); err != nil {
				log.Printf("Environment setup failed at 'make %s': %v", target, err)
				return
			}
		}
		log.Println("Environment setup finished")
	}()

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Environment setup started"))
}

// withCORS rejects requests from origins that are not allowed and sets the
// CORS headers for the ones that are. Requests without an Origin header
// (curl, same-machine tools) pass through.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			if !originAllowed(r) {
				http.Error(w, "origin not allowed", http.StatusForbidden)
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// evaluationTimeout bounds how long one request may spend querying the
// cluster for all questions.
var evaluationTimeout = 20 * time.Second

// evaluate checks every question, giving up when the client disconnects or
// evaluationTimeout passes. On failure it writes the error response and
// returns false.
//
// Two optional query parameters limit which questions are evaluated: "track"
// (core, cks, ckne or all) and "difficulty" (easy, medium, hard or all).
func evaluate(w http.ResponseWriter, r *http.Request, clients resources.Clients) ([]utils.Result, bool) {
	track, ok := resources.ParseTrack(r.URL.Query().Get("track"))
	if !ok {
		http.Error(w, "unknown track: use core, cks, ckne or all", http.StatusBadRequest)
		return nil, false
	}

	difficulty, ok := resources.ParseDifficulty(r.URL.Query().Get("difficulty"))
	if !ok {
		http.Error(w, "unknown difficulty: use easy, medium, hard or all", http.StatusBadRequest)
		return nil, false
	}

	ctx, cancel := context.WithTimeout(r.Context(), evaluationTimeout)
	defer cancel()

	results, err := resources.Evaluate(ctx, clients, resources.Filter{Track: track, Difficulty: difficulty})
	if err == nil {
		return results, true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		log.Printf("Cluster evaluation timed out after %s", evaluationTimeout)
		http.Error(w, "timed out waiting for the Kubernetes cluster", http.StatusGatewayTimeout)
	} else {
		log.Printf("Cluster evaluation aborted: %v", err)
		http.Error(w, "cluster evaluation aborted", http.StatusServiceUnavailable)
	}
	return nil, false
}

// getQuestions returns every question with its current pass/fail state.
func getQuestions(w http.ResponseWriter, r *http.Request, clients resources.Clients) {
	results, ok := evaluate(w, r, clients)
	if !ok {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// getCatalog lists every question without consulting the cluster, so the
// frontend can show them before an attempt starts.
func getCatalog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources.Catalog())
}

func startQuiz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Quiz started"))
}

// finishQuiz re-evaluates the cluster and returns the score together with the
// per-question results it was computed from.
func finishQuiz(w http.ResponseWriter, r *http.Request, clients resources.Clients) {
	results, ok := evaluate(w, r, clients)
	if !ok {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"score":   resources.Score(results),
		"results": results,
	})
}

// runTerminalCommand runs one shell command line and returns everything it
// printed (stdout and stderr). When the command fails, the failure reason is
// appended after its output, so the user still sees the command's own error
// message.
func runTerminalCommand(command string) []byte {
	output, err := exec.Command("sh", "-c", command).CombinedOutput()
	if err != nil {
		if len(output) > 0 && output[len(output)-1] != '\n' {
			output = append(output, '\n')
		}
		output = append(output, []byte(fmt.Sprintf("[%v]\n", err))...)
	}
	return output
}

// WebSocket terminal handler
func handleWebSocketTerminal(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	var buffer strings.Builder

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		buffer.WriteString(string(msg))

		// Quando o buffer recebe um newline, processa o comando
		if strings.Contains(buffer.String(), "\n") {
			conn.WriteMessage(websocket.TextMessage, runTerminalCommand(buffer.String()))
			buffer.Reset()
		}
	}
}

func main() {
	config, err := k8s.LoadKubeConfig()
	if err != nil {
		log.Fatalf("Failed to load Kubernetes configuration: %v", err)
	}
	clientset, err := k8s.NewClientSet(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes clientset: %v", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes dynamic client: %v", err)
	}
	clients := resources.Clients{Kube: clientset, Dynamic: dynamicClient}

	http.HandleFunc("/setup", setupEnvironment)
	http.HandleFunc("/catalog", getCatalog)
	http.HandleFunc("/questions", func(w http.ResponseWriter, r *http.Request) {
		getQuestions(w, r, clients)
	})
	http.HandleFunc("/start", startQuiz)
	http.HandleFunc("/finish", func(w http.ResponseWriter, r *http.Request) {
		finishQuiz(w, r, clients)
	})

	// WebSocket endpoint for terminal
	http.HandleFunc("/terminal", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocketTerminal(w, r)
	})

	// Start the server with CORS middleware applied globally
	log.Printf("KubeLearn backend listening on http://%s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, withCORS(http.DefaultServeMux)))
}

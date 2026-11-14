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
	"strings"
	"sync/atomic"
	"time"

	"kubelearn"
	"kubelearn/pkg/lab"
	"kubelearn/pkg/resources"
	"kubelearn/pkg/utils"

	"github.com/gorilla/websocket"
)

// listenAddr is loopback-only: the /terminal endpoint runs shell commands, so
// the server must never be reachable from other machines.
const listenAddr = "127.0.0.1:8083"

// allowedOrigins holds the browser origins allowed to call the API and open
// the terminal WebSocket: the interface served by this binary, plus the
// development server or the comma-separated list in KUBELEARN_ALLOWED_ORIGINS.
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
	// The interface embedded in this binary, under both names of the
	// loopback address.
	origins["http://"+listenAddr] = true
	origins["http://localhost"+listenAddr[strings.LastIndex(listenAddr, ":"):]] = true
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

// setupRunning guards against overlapping /setup runs.
var setupRunning atomic.Bool

// provisionLab creates the cluster and applies the lab resources. It is a
// variable so tests can replace it.
var provisionLab = func(ctx context.Context) error {
	return lab.New(kubelearn.Manifests(), log.Writer()).Up(ctx)
}

// setupEnvironment provisions the lab in the background. It only accepts
// POST, since it changes infrastructure.
func setupEnvironment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !setupRunning.CompareAndSwap(false, true) {
		http.Error(w, "environment setup already running", http.StatusConflict)
		return
	}

	go func() {
		defer setupRunning.Store(false)
		log.Println("Starting environment setup...")
		if err := provisionLab(context.Background()); err != nil {
			log.Printf("Environment setup failed: %v", err)
			return
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

// newMux wires the API routes and, at every other path, the web interface.
func newMux(clients resources.Clients) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/setup", setupEnvironment)
	mux.HandleFunc("/catalog", getCatalog)
	mux.HandleFunc("/questions", func(w http.ResponseWriter, r *http.Request) {
		getQuestions(w, r, clients)
	})
	mux.HandleFunc("/start", startQuiz)
	mux.HandleFunc("/finish", func(w http.ResponseWriter, r *http.Request) {
		finishQuiz(w, r, clients)
	})
	mux.HandleFunc("/terminal", handleWebSocketTerminal)
	mux.Handle("/", interfaceHandler())
	return mux
}

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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"kubelearn/pkg/k8s"
	"kubelearn/pkg/resources/easy"
	"kubelearn/pkg/resources/hard"
	"kubelearn/pkg/resources/medium"
	"kubelearn/pkg/utils"

	"github.com/gorilla/websocket"
	"k8s.io/client-go/kubernetes"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// runMakeCommand runs a Makefile target and logs the output.
func runMakeCommand(target string) error {
	cmd := exec.Command("make", target)
	cmd.Dir = "../"
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd.Run()
}

// setupEnvironment triggers the setup process using the Makefile.
func setupEnvironment(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	go func() {
		log.Println("Starting environment setup...")
		if err := runMakeCommand("init"); err != nil {
			log.Printf("Error initializing Terraform: %v", err)
			return
		}
		if err := runMakeCommand("apply"); err != nil {
			log.Printf("Error applying Terraform configurations: %v", err)
			return
		}
		if err := runMakeCommand("all"); err != nil {
			log.Printf("Error applying Kubernetes manifests: %v", err)
		}
	}()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Environment setup started"))
}

// CORS middleware
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func getQuestions(w http.ResponseWriter, clientset *kubernetes.Clientset) {
	questions := []utils.Result{
		easy.CreatePod(clientset),                       // Question 1 - Easy
		medium.CreateDeployment(clientset),              // Question 2 - Medium
		hard.CreateDeploymentAndService(clientset),      // Question 3 - Hard
		easy.CreateNamespace(clientset),                 // Question 4 - Easy
		medium.CreateConfigMap(clientset),               // Question 5 - Medium
		medium.CreateLabel(clientset),                   // Question 6 - Medium
		medium.CreatePersistentVolume(clientset),        // Question 7 - Medium
		medium.CreatePersistentVolumeClaim(clientset),   // Question 8 - Medium
		hard.CreatePodVolumeClaim(clientset),            // Question 9 - Hard
		hard.CheckPodError(clientset),                   // Question 10 - Hard
		hard.CreateNetPolRule(clientset),                // Question 11 - Hard
		easy.CreateSecret(clientset),                    // Question 12 - Easy
		hard.CreatePodAddSecret(clientset),              // Question 13 - Hard
		easy.CreateServiceAccount(clientset),            // Question 14 - Easy
		medium.AddServiceAccountToDeployment(clientset), // Question 15 - Medium
		medium.ChangeReplicaCount(clientset),            // Question 16 - Medium
		medium.CreateHpa(clientset),                     // Question 17 - Medium
		medium.AddSecurityContext(clientset),            // Question 18 - Medium
		medium.AddLivenessProbe(clientset),              // Question 19 - Medium
		easy.CreateDeploymentYellow(clientset),          // Question 20 - Easy
		hard.CreateServiceForYellow(clientset),          // Question 21 - Hard
		hard.CreateIngressYellow(clientset),             // Question 22 - Hard
		hard.CreateRoleOne(clientset),                   // Question 23 - Hard
		medium.CreateJob(clientset),                     // Question 24 - Medium
		medium.CreateCronjob(clientset),                 // Question 25 - Medium
		hard.CreateStatefulSet(clientset),               // Question 26 - Hard
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(questions)
}

func startQuiz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Quiz started"))
}

func finishQuiz(w http.ResponseWriter, clientset *kubernetes.Clientset) {
	questions := []utils.Result{
		easy.CreatePod(clientset),                       // Question 1 - Easy
		medium.CreateDeployment(clientset),              // Question 2 - Medium
		hard.CreateDeploymentAndService(clientset),      // Question 3 - Hard
		easy.CreateNamespace(clientset),                 // Question 4 - Easy
		medium.CreateConfigMap(clientset),               // Question 5 - Medium
		medium.CreateLabel(clientset),                   // Question 6 - Medium
		medium.CreatePersistentVolume(clientset),        // Question 7 - Medium
		medium.CreatePersistentVolumeClaim(clientset),   // Question 8 - Medium
		hard.CreatePodVolumeClaim(clientset),            // Question 9 - Hard
		hard.CheckPodError(clientset),                   // Question 10 - Hard
		hard.CreateNetPolRule(clientset),                // Question 11 - Hard
		easy.CreateSecret(clientset),                    // Question 12 - Easy
		hard.CreatePodAddSecret(clientset),              // Question 13 - Hard
		easy.CreateServiceAccount(clientset),            // Question 14 - Easy
		medium.AddServiceAccountToDeployment(clientset), // Question 15 - Medium
		medium.ChangeReplicaCount(clientset),            // Question 16 - Medium
		medium.CreateHpa(clientset),                     // Question 17 - Medium
		medium.AddSecurityContext(clientset),            // Question 18 - Medium
		medium.AddLivenessProbe(clientset),              // Question 19 - Medium
		easy.CreateDeploymentYellow(clientset),          // Question 20 - Easy
		hard.CreateServiceForYellow(clientset),          // Question 21 - Hard
		hard.CreateIngressYellow(clientset),             // Question 22 - Hard
		hard.CreateRoleOne(clientset),                   // Question 23 - Hard
		medium.CreateJob(clientset),                     // Question 24 - Medium
		medium.CreateCronjob(clientset),                 // Question 25 - Medium
		hard.CreateStatefulSet(clientset),               // Question 26 - Hard
	}

	correctAnswers := 0
	for _, question := range questions {
		if question.Passed {
			correctAnswers++
		}
	}
	totalQuestions := len(questions)
	score := (float64(correctAnswers) / float64(totalQuestions)) * 100

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"score": score,
	})
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
			cmd := exec.Command("sh", "-c", buffer.String())
			output, err := cmd.CombinedOutput()
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error: %v\n", err)))
			} else {
				conn.WriteMessage(websocket.TextMessage, output)
			}
			buffer.Reset()
		}
	}
}

func main() {
	config := k8s.LoadKubeConfig()
	clientset, err := k8s.NewClientSet(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes clientset: %v", err)
	}

	http.HandleFunc("/setup", setupEnvironment)
	http.HandleFunc("/questions", func(w http.ResponseWriter, r *http.Request) {
		getQuestions(w, clientset)
	})
	http.HandleFunc("/start", startQuiz)
	http.HandleFunc("/finish", func(w http.ResponseWriter, r *http.Request) {
		finishQuiz(w, clientset)
	})

	// WebSocket endpoint for terminal
	http.HandleFunc("/terminal", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocketTerminal(w, r)
	})

	// Start the server with CORS middleware applied globally
	log.Fatal(http.ListenAndServe(":8083", withCORS(http.DefaultServeMux)))
}

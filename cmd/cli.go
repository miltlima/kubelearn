package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"

	"kubelearn"
	"kubelearn/pkg/k8s"
	"kubelearn/pkg/lab"
	"kubelearn/pkg/resources"

	"k8s.io/client-go/dynamic"
)

// version is set at build time (-ldflags "-X main.version=...").
var version = "dev"

const usage = `KubeLearn - hands-on Kubernetes practice on a local cluster.

Usage:
  kubelearn up        Create the lab cluster (kind) and apply the lab resources
  kubelearn           Start KubeLearn and open it in the browser
  kubelearn down      Delete the lab cluster
  kubelearn version   Print the version

Options for "kubelearn" (serve):
  -open=false         Do not open the browser

Requirements: Docker, kind and kubectl.
`

func main() {
	args := os.Args[1:]
	command := "serve"
	if len(args) > 0 && len(args[0]) > 0 && args[0][0] != '-' {
		command, args = args[0], args[1:]
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var err error
	switch command {
	case "serve":
		err = serve(args)
	case "up":
		err = lab.New(kubelearn.Manifests(), os.Stdout).Up(ctx)
		if err == nil {
			fmt.Println("Run \"kubelearn\" to start practising.")
		}
	case "down":
		err = lab.New(kubelearn.Manifests(), os.Stdout).Down(ctx)
	case "version":
		fmt.Println(version)
	case "help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", command, usage)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// serve starts the API, the terminal and the web interface.
func serve(args []string) error {
	flags := flag.NewFlagSet("kubelearn", flag.ContinueOnError)
	flags.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	open := flags.Bool("open", true, "open the browser")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	config, err := k8s.LoadKubeConfig()
	if err != nil {
		return fmt.Errorf("no Kubernetes cluster configured (%v).\nRun \"kubelearn up\" to create the lab cluster first", err)
	}
	clientset, err := k8s.NewClientSet(config)
	if err != nil {
		return fmt.Errorf("creating the Kubernetes client: %w", err)
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("creating the Kubernetes dynamic client: %w", err)
	}
	clients := resources.Clients{Kube: clientset, Dynamic: dynamicClient}

	// Answers are checked, and terminal commands run, against the current
	// kubectl context. Say so when it is not the lab cluster.
	if current := k8s.CurrentContext(); current != lab.Context {
		log.Printf("Warning: the current kubectl context is %q, not the lab cluster %q. "+
			"Questions are checked and terminal commands run against %q. "+
			"Run \"kubelearn up\" to create and select the lab cluster.", current, lab.Context, current)
	}

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("cannot listen on %s (is KubeLearn already running?): %w", listenAddr, err)
	}
	url := "http://" + listenAddr
	log.Printf("KubeLearn %s listening on %s", version, url)

	if _, embedded := kubelearn.Frontend(); embedded && *open {
		openBrowser(url)
	}
	return http.Serve(listener, withCORS(newMux(clients)))
}

// openBrowser opens url in the default browser. Failing to do so is not an
// error: the address is printed above.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("Open %s in your browser.", url)
	}
}

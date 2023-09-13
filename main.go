package main

import (
	"context"
	"log"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

var kubeconfig = os.Getenv("HOME") + "/.kube/config"

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error to build kubeconfig: %s", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error to create clientset Kubernetes: %s", err)
	}

	results := []Result{
		testPod(clientset),
		testDeployment(clientset),
		testDeploymentAndService(clientset),
	}

	renderResultsTable(results)
}

type Result struct {
	TestName string
	Passed   bool
}

func testPod(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace = "default"
		expectedPodName   = "nginx"
	)
	pod, err := clientset.CoreV1().Pods(expectedNamespace).Get(context.TODO(), expectedPodName, metav1.GetOptions{})
	passed := err == nil && pod.Spec.Containers[0].Image == "nginx:alpine" && pod.Name == "nginx"
	return Result{
		TestName: "Question 1 - Deploy a POD with nginx:alpine image",
		Passed:   passed,
	}
}

func testDeployment(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace      = "default"
		expectedDeploymentName = "nginx-deployment"
		expectedReplicas       = int32(4)
		expectedImage          = "nginx:alpine"
	)
	deployment, err := clientset.AppsV1().Deployments(expectedNamespace).Get(context.TODO(), expectedDeploymentName, metav1.GetOptions{})
	passed := err == nil && expectedDeploymentName == deployment.Name && expectedReplicas == *deployment.Spec.Replicas && expectedImage == deployment.Spec.Template.Spec.Containers[0].Image
	return Result{
		TestName: "Question 2 - Create a deployment with nginx:alpine image and 4 replicas",
		Passed:   passed,
	}
}

func testDeploymentAndService(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace      = "latam"
		expectedDeploymentName = "redis"
		expectedServiceName    = "redis-service"
		expectedServicePort    = int32(6379)
		expectedImage          = "redis:alpine"
	)

	deployment, err := clientset.AppsV1().Deployments(expectedNamespace).Get(context.TODO(), expectedDeploymentName, metav1.GetOptions{})
	service, err := clientset.CoreV1().Services(expectedNamespace).Get(context.TODO(), expectedServiceName, metav1.GetOptions{})
	passed := err == nil && expectedDeploymentName == deployment.Name && expectedServiceName == service.Name && expectedServicePort == service.Spec.Ports[0].Port && expectedImage == deployment.Spec.Template.Spec.Containers[0].Image
	return Result{
		TestName: "Question 3 - Create a deployment with redis:alpine image and a service with port 6379 in namespace latam",
		Passed:   passed,
	}
}

func renderResultsTable(results []Result) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Test your knowledge of Kubernetes", "Result"})

	table.SetAutoWrapText(false)

	for _, result := range results {
		passedStr := color.GreenString("Passed")
		if !result.Passed {
			passedStr = color.RedString("Failed")
		}
		row := []string{result.TestName, passedStr}
		table.Append(row)
	}

	table.Render()
}

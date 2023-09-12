package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig string
	logFile    *os.File
)

func TestQuestionOne(t *testing.T) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %s", err)
	}

	namespace := "default"
	podName := "nginx"
	expectImage := "nginx:alpine"

	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Error to get pod %s in namespace %s: %v", podName, namespace, err)
	}

	actualImage := pod.Spec.Containers[0].Image
	assert.Equal(t, expectImage, actualImage, "Image should be %s", expectImage)
}

func TestQuestionTwo(t *testing.T) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %s", err)
	}

	namespace := "default"
	expectedDeploymentName := "nginx-deployment"
	expectedReplicas := int32(4)
	expectedImage := "nginx:alpine"

	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), expectedDeploymentName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Error to get deployment %s in namespace %s: %v", expectedDeploymentName, namespace, err)
	}

	actualReplicas := *deployment.Spec.Replicas
	actualImage := deployment.Spec.Template.Spec.Containers[0].Image
	actualDeploymentName := deployment.Name

	assert.Equal(t, expectedDeploymentName, actualDeploymentName, "Deployment name should be %s", expectedDeploymentName)
	assert.Equal(t, expectedReplicas, actualReplicas, "Replicas should be %d", expectedReplicas)
	assert.Equal(t, expectedImage, actualImage, "Image should be %s", expectedImage)

}

func TestQuestionThree(t *testing.T) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %s", err)
	}

	expectedNamespace := "latam"
	expectedDeploymentName := "redis"
	expectedServiceName := "redis-service"
	expectedServicePort := int32(6379)
	expectedImage := "redis:alpine"

	deployment, err := clientset.AppsV1().Deployments(expectedNamespace).Get(context.TODO(), expectedDeploymentName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Error to get deployment %s in namespace %s: %v", expectedDeploymentName, expectedNamespace, err)
	}

	service, err := clientset.CoreV1().Services(expectedNamespace).Get(context.TODO(), expectedServiceName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Error to get service %s in namespace %s: %v", expectedServiceName, expectedNamespace, err)
	}

	actualNamespace := service.Namespace
	actualDeploymentName := deployment.Name
	actualImage := deployment.Spec.Template.Spec.Containers[0].Image

	assert.Equal(t, expectedServiceName, service.Name, "Service name should be %s", expectedServiceName)
	assert.Equal(t, expectedServicePort, service.Spec.Ports[0].Port, "Service port should be %d", expectedServicePort)
	assert.Equal(t, expectedDeploymentName, actualDeploymentName, "Deployment name should be %s", expectedDeploymentName)
	assert.Equal(t, expectedNamespace, actualNamespace, "Service namespace should be %s", expectedNamespace)
	assert.Equal(t, expectedImage, actualImage, "Image should be %s", expectedImage)

}
func TestMain(m *testing.M) {
	flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.Parse()

	var err error
	logFile, err = os.OpenFile("test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("failed to open log file", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	os.Exit(m.Run())
}

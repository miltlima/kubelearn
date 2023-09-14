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
		testNamespace(clientset),
		testConfigMap(clientset),
		testLabel(clientset),
		testPersistentVolume(clientset),
		testPersistentVolumeClaim(clientset),
	}

	renderResultsTable(results)
}

type Result struct {
	TestName   string
	Passed     bool
	Difficulty string
}

func testPod(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace = "default"
		expectedPodName   = "nginx"
		expectedImage     = "nginx:alpine"
	)
	pod, err := clientset.CoreV1().Pods(expectedNamespace).Get(context.TODO(), expectedPodName, metav1.GetOptions{})
	passed := err == nil && pod.Spec.Containers[0].Image == expectedImage && pod.Name == expectedPodName
	return Result{
		TestName:   "Question 1 - Create a pod nginx name with nginx:alpine image",
		Passed:     passed,
		Difficulty: "Easy",
	}
}

func testDeployment(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace      = "default"
		expectedDeploymentName = "nginx-deployment"
		expectedReplicas       = int32(4)
		expectedImage          = "nginx:1.17"
	)
	deployment, err := clientset.AppsV1().Deployments(expectedNamespace).Get(context.TODO(), expectedDeploymentName, metav1.GetOptions{})
	passed := err == nil && expectedDeploymentName == deployment.Name && expectedReplicas == *deployment.Spec.Replicas && expectedImage == deployment.Spec.Template.Spec.Containers[0].Image
	return Result{
		TestName:   "Question 2 - Create a deployment nginx-deployment with nginx:alpine image and 4 replicas",
		Passed:     passed,
		Difficulty: "Medium",
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
		TestName:   "Question 3 - Create a deployment redis name with redis:alpine image and a service with port 6379 in namespace latam",
		Passed:     passed,
		Difficulty: "Hard",
	}
}

func testNamespace(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace = "emea"
	)
	namespace, err := clientset.CoreV1().Namespaces().Get(context.TODO(), expectedNamespace, metav1.GetOptions{})
	passed := err == nil && expectedNamespace == namespace.Name
	return Result{
		TestName:   "Question 4 - Create a namespace europe",
		Passed:     passed,
		Difficulty: "Easy",
	}
}

func testConfigMap(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace     = "default"
		expectedConfigMapName = "europe-configmap"
		expectedDataKey       = "France"
		expectedDataValue     = "Paris"
	)
	configMap, err := clientset.CoreV1().ConfigMaps(expectedNamespace).Get(context.TODO(), expectedConfigMapName, metav1.GetOptions{})
	passed := err == nil && expectedConfigMapName == configMap.Name && expectedDataValue == configMap.Data[expectedDataKey]
	return Result{
		TestName:   "Question 5 - Create a configmap europe-configmap with data France=Paris",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

func testLabel(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace  = "asia"
		expectedPodName    = "tshoot"
		expectedImage      = "amazon/amazon-ecs-network-sidecar:latest"
		expectedLabelKey   = "country"
		expectedLabelValue = "china"
	)
	pod, err := clientset.CoreV1().Pods(expectedNamespace).Get(context.TODO(), expectedPodName, metav1.GetOptions{})
	passed := err == nil && expectedImage == pod.Spec.Containers[0].Image && expectedLabelValue == pod.ObjectMeta.Labels[expectedLabelKey]

	return Result{
		TestName:   "Question 6 - Create a pod thsoot with label country=china with amazon/amazon-ecs-network-sidecar:latest image and namespace asia",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

func testPersistentVolume(clientset *kubernetes.Clientset) Result {
	const (
		expectedPersistentVolumeName = "unicorn-pv"
		expectedCapacity             = "1Gi"
		expectedAccessMode           = "ReadWriteMany"
		expectedHostPath             = "/tmp/data"
	)
	pv, err := clientset.CoreV1().PersistentVolumes().Get(context.TODO(), expectedPersistentVolumeName, metav1.GetOptions{})
	passed := err == nil && expectedCapacity == pv.Spec.Capacity.Storage().String() && expectedAccessMode == pv.Spec.AccessModes[0] && expectedHostPath == pv.Spec.HostPath.Path

	return Result{
		TestName:   "Question 7 - Create a persistent volume unicorn-pv with capacity 1Gi and access mode ReadWriteMany and host path /tmp/data",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

func testPersistentVolumeClaim(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace                 = "default"
		expectedPersistentVolumeClaimName = "unicorn-pvc"
		expectedAccessMode                = "ReadWriteMany"
		expectedCapacity                  = "400Mi"
	)
	pvc, err := clientset.CoreV1().PersistentVolumeClaims(expectedNamespace).Get(context.TODO(), expectedPersistentVolumeClaimName, metav1.GetOptions{})
	passed := err == nil && expectedCapacity == pvc.Spec.Resources.Requests.Storage().String() && expectedAccessMode == pvc.Spec.AccessModes[0]

	return Result{
		TestName:   "Question 8 - Create a persistent volume claim unicorn-pvc with capacity 400Mi and access mode ReadWriteMany",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

func renderResultsTable(results []Result) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"KubeLearn - Test your knowledge of Kubernetes", "Result", "Difficulty"})

	table.SetAutoWrapText(false)

	for _, result := range results {
		passedStr := color.GreenString("✅ Pass")
		if !result.Passed {
			passedStr = color.RedString("🆘 Fail")
		}
		row := []string{result.TestName, passedStr, result.Difficulty}
		table.Append(row)
	}

	table.Render()
}

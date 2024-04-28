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
	"log"
	"os"

	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
		createPod(clientset),
		createDeployment(clientset),
		createDeploymentAndService(clientset),
		createNamespace(clientset),
		createConfigMap(clientset),
		createLabel(clientset),
		createPersistentVolume(clientset),
		createPersistentVolumeClaim(clientset),
		createPodVolumeClaim(clientset),
		checkPodError(clientset),
		createNetPolRule(clientset),
		createSecret(clientset),
		createPodAddSecret(clientset),
		createServiceAccount(clientset),
		addServiceAccountToDeployment(clientset),
		changeReplicaCount(clientset),
		createHpa(clientset),
		addSecurityContext(clientset),
		addLivenessProbe(clientset),
		createDeploymentYellow(clientset),
		createServiceForYellow(clientset),
		createIngressYellow(clientset),
		createRoleOne(clientset),
		createJob(clientset),
		createCronjob(clientset),
		createStateFulSet(clientset),
	}

	renderResultsTable(results)
}

type Result struct {
	TestName   string
	Passed     bool
	Difficulty string
}

func createPod(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace = "default"
		expectedPodName   = "nginx"
		expectedImage     = "nginx:alpine"
	)

	pod, err := clientset.CoreV1().Pods(expectedNamespace).Get(context.TODO(), expectedPodName, metav1.GetOptions{})

	passed := err == nil &&
		pod.Spec.Containers[0].Image == expectedImage &&
		pod.Name == expectedPodName

	return Result{
		TestName:   "Question 1 - Create a pod nginx name with nginx:alpine image",
		Passed:     passed,
		Difficulty: "Easy",
	}
}

func createDeployment(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace      = "default"
		expectedDeploymentName = "nginx-deployment"
		expectedReplicas       = int32(4)
		expectedImage          = "nginx:alpine"
	)

	deployment, err := clientset.AppsV1().Deployments(expectedNamespace).Get(context.TODO(), expectedDeploymentName, metav1.GetOptions{})

	passed := err == nil &&
		expectedDeploymentName == deployment.Name &&
		expectedReplicas == *deployment.Spec.Replicas &&
		expectedImage == deployment.Spec.Template.Spec.Containers[0].Image

	return Result{
		TestName:   "Question 2 - Create a deployment nginx-deployment with nginx:alpine image and 4 replicas",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

func createDeploymentAndService(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace      = "latam"
		expectedDeploymentName = "redis"
		expectedServiceName    = "redis-service"
		expectedServicePort    = int32(6379)
		expectedImage          = "redis:alpine"
	)

	deployment, err := clientset.AppsV1().Deployments(expectedNamespace).Get(context.TODO(), expectedDeploymentName, metav1.GetOptions{})
	service, err := clientset.CoreV1().Services(expectedNamespace).Get(context.TODO(), expectedServiceName, metav1.GetOptions{})

	passed := err == nil &&
		service != nil &&
		expectedDeploymentName == deployment.Name &&
		expectedServiceName == service.Name &&
		expectedServicePort == service.Spec.Ports[0].Port &&
		expectedImage == deployment.Spec.Template.Spec.Containers[0].Image

	return Result{
		TestName:   "Question 3 - Create a deployment redis name with redis:alpine image and a service with name redis-service and port 6379 in namespace latam",
		Passed:     passed,
		Difficulty: "Hard",
	}
}

func createNamespace(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace = "europe"
	)

	namespace, err := clientset.CoreV1().Namespaces().Get(context.TODO(), expectedNamespace, metav1.GetOptions{})

	passed := err == nil &&
		expectedNamespace == namespace.Name

	return Result{
		TestName:   "Question 4 - Create a namespace europe",
		Passed:     passed,
		Difficulty: "Easy",
	}
}

func createConfigMap(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace     = "default"
		expectedConfigMapName = "europe-configmap"
		expectedDataKey       = "France"
		expectedDataValue     = "Paris"
	)

	configMap, err := clientset.CoreV1().ConfigMaps(expectedNamespace).Get(context.TODO(), expectedConfigMapName, metav1.GetOptions{})

	passed := err == nil &&
		expectedConfigMapName == configMap.Name &&
		expectedDataValue == configMap.Data[expectedDataKey]

	return Result{
		TestName:   "Question 5 - Create a configmap europe-configmap with data France=Paris",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

func createLabel(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace  = "asia"
		expectedPodName    = "tshoot"
		expectedImage      = "amazon/amazon-ecs-network-sidecar:latest"
		expectedLabelKey   = "country"
		expectedLabelValue = "china"
	)

	pod, err := clientset.CoreV1().Pods(expectedNamespace).Get(context.TODO(), expectedPodName, metav1.GetOptions{})

	passed := err == nil &&
		expectedImage == pod.Spec.Containers[0].Image &&
		expectedLabelValue == pod.ObjectMeta.Labels[expectedLabelKey]

	return Result{
		TestName:   "Question 6 - Create a pod thsoot with label country=china with amazon/amazon-ecs-network-sidecar:latest image and namespace asia",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

func createPersistentVolume(clientset *kubernetes.Clientset) Result {
	const (
		expectedPersistentVolumeName = "unicorn-pv"
		expectedCapacity             = "1Gi"
		expectedAccessMode           = "ReadWriteMany"
		expectedHostPath             = "/tmp/data"
	)

	pv, err := clientset.CoreV1().PersistentVolumes().Get(context.TODO(), expectedPersistentVolumeName, metav1.GetOptions{})

	passed := err == nil &&
		expectedCapacity == pv.Spec.Capacity.Storage().String() &&
		expectedAccessMode == pv.Spec.AccessModes[0] &&
		expectedHostPath == pv.Spec.HostPath.Path

	return Result{
		TestName:   "Question 7 - Create a persistent volume unicorn-pv with capacity 1Gi and access mode ReadWriteMany and host path /tmp/data",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

func createPersistentVolumeClaim(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace                 = "default"
		expectedPersistentVolumeClaimName = "unicorn-pvc"
		expectedAccessMode                = "ReadWriteMany"
		expectedCapacity                  = "400Mi"
	)

	pvc, err := clientset.CoreV1().PersistentVolumeClaims(expectedNamespace).Get(context.TODO(), expectedPersistentVolumeClaimName, metav1.GetOptions{})

	passed := err == nil &&
		expectedCapacity == pvc.Spec.Resources.Requests.Storage().String() &&
		expectedAccessMode == pvc.Spec.AccessModes[0]

	return Result{
		TestName:   "Question 8 - Create a persistent volume claim unicorn-pvc with capacity 400Mi and access mode ReadWriteMany",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

func createPodVolumeClaim(clientset *kubernetes.Clientset) Result {
	const (
		expectedPodName               = "webserver"
		expectedNamespace             = "public"
		expectedVolumeName            = "unicorn-pv"
		expectedPersistentVolumeClaim = "unicorn-pvc"
		expectedImage                 = "nginx:alpine"
		expectedVolumeMount           = "/usr/share/nginx/html"
	)

	pod, err := clientset.CoreV1().Pods(expectedNamespace).Get(context.TODO(), expectedPodName, metav1.GetOptions{})

	passed := err == nil &&
		expectedImage == pod.Spec.Containers[0].Image &&
		expectedPersistentVolumeClaim == pod.Spec.Volumes[0].VolumeSource.PersistentVolumeClaim.ClaimName &&
		expectedVolumeMount == pod.Spec.Containers[0].VolumeMounts[0].MountPath &&
		expectedVolumeName == pod.Spec.Volumes[0].Name

	return Result{
		TestName:   "Question 9 - Create a pod webserver in public namespace with nginx:alpine image and a volume mount /usr/share/nginx/html and a persistent volume claim unicorn-pvc",
		Passed:     passed,
		Difficulty: "Hard",
	}
}

func checkPodError(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace = "bandai"
		expectedPodName   = "gundamv"
		expectedImage     = "nginx:alpine"
	)

	pod, err := clientset.CoreV1().Pods(expectedNamespace).Get(context.TODO(), expectedPodName, metav1.GetOptions{})

	passed := err == nil &&
		expectedImage == pod.Spec.Containers[0].Image

	return Result{
		TestName:   "Question 10 - There is a pod with problem, Can you able to solve it ? Find the problem and fix it",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

func createNetPolRule(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace  = "colors"
		expectedNetPolName = "allow-policy-colors"
		expectedFromLabel  = "tier=frontend"
		expectedToLabel    = "tier=backend"
		expectedPort       = int32(6379)
	)

	netPol, err := clientset.NetworkingV1().NetworkPolicies(expectedNamespace).Get(context.TODO(), expectedNetPolName, metav1.GetOptions{})

	passed := err == nil && hasCorrectIngressRule(netPol.Spec.Ingress)

	return Result{
		TestName:   "Question 11 - Create a network policy allow-policy-colors to allow redmobile-webserver to access bluemobile-dbcache.",
		Passed:     passed,
		Difficulty: "Hard",
	}
}

func createSecret(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace  = "colors"
		expectedSecretName = "secret-colors"
		expectedDataKey    = "color"
		expectedDataValue  = "red"
	)

	secret, err := clientset.CoreV1().Secrets(expectedNamespace).Get(context.TODO(), expectedSecretName, metav1.GetOptions{})

	passed := err == nil &&
		expectedSecretName == secret.Name &&
		expectedDataValue == string(secret.Data[expectedDataKey])

	return Result{
		TestName:   "Question 12 - Create a secret secret-colors with data color=red in colors namespace",
		Passed:     passed,
		Difficulty: "Easy",
	}
}

func createPodAddSecret(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace  = "colors"
		expectedSecretName = "secret-purple"
		expectedPodName    = "purple"
		expectedImage      = "redis:alpine"
		expectedDataKey    = "singer"
		expectedDataValue  = "prince"
	)

	pod, err := clientset.CoreV1().Pods(expectedNamespace).Get(context.TODO(), expectedPodName, metav1.GetOptions{})
	secret, err := clientset.CoreV1().Secrets(expectedNamespace).Get(context.TODO(), expectedSecretName, metav1.GetOptions{})

	passed := err == nil &&
		expectedSecretName == pod.Spec.Volumes[0].Secret.SecretName &&
		expectedDataValue == string(secret.Data[expectedDataKey]) &&
		expectedImage == pod.Spec.Containers[0].Image

	return Result{
		TestName:   "Question 13 - Add a secret secret-purple with data singer=prince to the pod purple with image redis:alpine in colors namespace",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

func createServiceAccount(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace          = "default"
		expectedServiceAccountName = "america-sa"
	)

	sa, err := clientset.CoreV1().ServiceAccounts(expectedNamespace).Get(context.TODO(), expectedServiceAccountName, metav1.GetOptions{})

	if err != nil {
		return Result{
			TestName:   "Question 14 - Create a service account america-sa in default namespace",
			Passed:     false,
			Difficulty: "Easy",
		}
	}

	passed := err == nil &&
		expectedServiceAccountName == sa.Name

	return Result{
		TestName:   "Question 14 - Create a service account america-sa in default namespace",
		Passed:     passed,
		Difficulty: "Easy",
	}
}

func addServiceAccountToDeployment(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace          = "default"
		expectedDeploymentName     = "mark42"
		expectedServiceAccountName = "america-sa"
	)

	deploy, err := clientset.AppsV1().Deployments(expectedNamespace).Get(context.TODO(), expectedDeploymentName, metav1.GetOptions{})

	if err != nil {
		return Result{
			TestName:   "Question 15 - Add service account america-sa to the deployment mark42",
			Passed:     false,
			Difficulty: "Easy",
		}
	}

	passed := err == nil &&
		expectedServiceAccountName == deploy.Spec.Template.Spec.ServiceAccountName

	return Result{
		TestName:   "Question 15 - Add service account america-sa to the deployment mark42",
		Passed:     passed,
		Difficulty: "Easy",
	}
}

func changeReplicaCount(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace      = "default"
		expectedDeploymentName = "mark42"
		expectedReplicas       = int32(5)
	)

	deploy, err := clientset.AppsV1().Deployments(expectedNamespace).Get(context.TODO(), expectedDeploymentName, metav1.GetOptions{})

	if err != nil {
		return Result{
			TestName:   "Question 16 - Change the replica count of the deployment mark42 to 5",
			Passed:     false,
			Difficulty: "Easy",
		}
	}

	passed := err == nil &&
		expectedReplicas == *deploy.Spec.Replicas

	return Result{
		TestName:   "Question 16 - Change the replica count of the deployment mark42 to 5",
		Passed:     passed,
		Difficulty: "Easy",
	}
}

func createHpa(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace      = "default"
		expectedDeploymentName = "mark43"
		expectedMinReplicas    = int32(2)
		expectedMaxReplicas    = int32(8)
		expectedCpuPercent     = int32(80)
	)

	hpa, err := clientset.AutoscalingV2().HorizontalPodAutoscalers(expectedNamespace).Get(context.TODO(), expectedDeploymentName, metav1.GetOptions{})

	if err != nil {
		return Result{
			TestName:   "Question 17 - Create a horizontal pod autoscaler hpa-mark43 for deployment mark43 with cpu percent 80, min replicas 2 and max replicas 8",
			Passed:     false,
			Difficulty: "Medium",
		}
	}

	passed := err == nil &&
		expectedDeploymentName == hpa.Spec.ScaleTargetRef.Name &&
		expectedMinReplicas == *hpa.Spec.MinReplicas &&
		expectedMaxReplicas == hpa.Spec.MaxReplicas &&
		expectedCpuPercent == *hpa.Spec.Metrics[0].Resource.Target.AverageUtilization

	return Result{
		TestName:   "Question 17 - Create a horizontal pod autoscaler hpa-mark43 for deployment mark43 with cpu percent 80, min replicas 2 and max replicas 8",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

func addSecurityContext(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace           = "default"
		expectedDeploymentName      = "mark42"
		expectedPrivilegeEscalation = false
	)

	deploy, err := clientset.AppsV1().Deployments(expectedNamespace).Get(context.TODO(), expectedDeploymentName, metav1.GetOptions{})

	if err != nil {
		return Result{
			TestName:   "Question 18 - Prevent privilege escalation in the deployment mark42",
			Passed:     false,
			Difficulty: "Medium",
		}
	}

	passed := deploy.Spec.Template.Spec.Containers[0].SecurityContext != nil &&
		*deploy.Spec.Template.Spec.Containers[0].SecurityContext.AllowPrivilegeEscalation == expectedPrivilegeEscalation

	return Result{
		TestName:   "Question 18 - Prevent privilege escalation in the deployment mark42",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

func addLivenessProbe(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace           = "shield"
		expectedPodName             = "mark50"
		expectedInitialDelaySeconds = int32(5)
		expectedPeriodSeconds       = int32(10)
		expectedLivenessProbeType   = "HttpGet"
		expectedLivenessProbePath   = "/"
		expectedLivenessProbePort   = int32(80)
	)

	pod, err := clientset.CoreV1().Pods(expectedNamespace).Get(context.TODO(), expectedPodName, metav1.GetOptions{})

	if err != nil || len(pod.Spec.Containers) == 0 || pod.Spec.Containers[0].LivenessProbe == nil {
		return Result{
			TestName:   "Question 19 - Add a liveness probe to the pod mark50 with initial delay 5s, period 10s HttpGet, port 80 and path '/' in namespace shield",
			Passed:     false,
			Difficulty: "Medium",
		}
	}

	passed := err == nil &&
		expectedInitialDelaySeconds == pod.Spec.Containers[0].LivenessProbe.InitialDelaySeconds &&
		expectedPeriodSeconds == pod.Spec.Containers[0].LivenessProbe.PeriodSeconds &&
		expectedLivenessProbePath == pod.Spec.Containers[0].LivenessProbe.HTTPGet.Path &&
		expectedLivenessProbePort == pod.Spec.Containers[0].LivenessProbe.HTTPGet.Port.IntVal

	return Result{
		TestName:   "Question 19 - Add a liveness probe to the pod mark50 with initial delay 5s, period 10s HttpGet, port 80 and path '/' in namespace shield",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

func createDeploymentYellow(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace = "colors"
		expectedName      = "yellow-deployment"
		expectedReplicas  = int32(2)
		expectedImage     = "bonovoo/node-app:1.0"
	)

	deployment, err := clientset.AppsV1().Deployments(expectedNamespace).Get(context.TODO(), expectedName, metav1.GetOptions{})

	// if err != nil {
	// 	return Result{
	// 		TestName:   "Question 20 - Create a deployment yellow-deployment with bonovoo/node-app:1.0 image and 2 replicas in namespace colors",
	// 		Passed:     false,
	// 		Difficulty: "Easy",
	// 	}
	// }

	passed := err == nil &&
		expectedImage == deployment.Spec.Template.Spec.Containers[0].Image &&
		expectedReplicas == *deployment.Spec.Replicas

	return Result{
		TestName:   "Question 20 - Create a deployment yellow-deployment with bonovoo/node-app:1.0 image and 2 replicas in namespace colors",
		Passed:     passed,
		Difficulty: "Easy",
	}
}

func createServiceForYellow(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace    = "colors"
		expectedServiceName  = "yellow-service"
		expectedPort         = 80
		expectedTargetPort   = int32(3000)
		expectedTargetObject = "yellow-deployment"
	)

	service, err := clientset.CoreV1().Services(expectedNamespace).Get(context.TODO(), expectedServiceName, metav1.GetOptions{})

	if err != nil {
		return Result{
			TestName:   "Question 21 - Create a service yellow-service for the deployment yellow-deployment in namespace colors with port 80 and target port 3000",
			Passed:     false,
			Difficulty: "Hard",
		}
	}

	passed := err == nil &&
		expectedPort == service.Spec.Ports[0].Port &&
		expectedTargetPort == service.Spec.Ports[0].TargetPort.IntVal

	for key, value := range service.Spec.Selector {
		if key == "app" && value == "yellow-deployment" {
			passed = true
			break
		}
	}

	return Result{
		TestName:   "Question 21 - Create a service yellow-service for the deployment yellow-deployment in namespace colors with port 80 and target port 3000",
		Passed:     passed,
		Difficulty: "Hard",
	}
}

func createIngressYellow(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace = "colors"
		expectedName      = "ingress-colors"
		expectedHost      = "yellow.com"
		expectedPath      = "/yellow"
		expectedService   = "yellow-service"
	)

	ingress, err := clientset.NetworkingV1().Ingresses(expectedNamespace).Get(context.TODO(), expectedName, metav1.GetOptions{})

	if err != nil {
		return Result{
			TestName:   "Question 22 - Create an ingress ingress-colors with host yellow.com, path /yellow and service yellow-service in namespace colors",
			Passed:     false,
			Difficulty: "Hard",
		}
	}

	ingressSpec := &ingress.Spec
	rules := ingressSpec.Rules

	passed := len(rules) > 0 &&
		expectedHost == rules[0].Host &&
		len(rules[0].HTTP.Paths) > 0 &&
		expectedPath == rules[0].HTTP.Paths[0].Path &&
		expectedService == rules[0].HTTP.Paths[0].Backend.Service.Name

	return Result{
		TestName:   "Question 22 - Create an ingress ingress-colors with host yellow.com, path /yellow and service yellow-service in namespace colors",
		Passed:     passed,
		Difficulty: "Hard",
	}
}

func createRoleOne(clientset *kubernetes.Clientset) Result {
	const (
		expectedName      = "apple-one"
		expectedNamespace = "fruits"
		expectedResource  = "pods"
	)

	expectedVerbs := []string{"get", "list", "watch"}

	role, err := clientset.RbacV1().Roles(expectedNamespace).Get(context.TODO(), expectedName, metav1.GetOptions{})

	if err != nil {
		return Result{
			TestName:   "Question 23 - Create a role apple-one with verbs get, list, watch in namespace fruits",
			Passed:     false,
			Difficulty: "Medium",
		}
	}

	passed := err == nil &&
		expectedResource == role.Rules[0].Resources[0]

	for _, verb := range expectedVerbs {
		if !contains(role.Rules[0].Verbs, verb) {
			passed = false
			break
		}
	}

	return Result{
		TestName:   "Question 23 - Create a role apple-one with verbs get, list, watch in namespace fruits",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

func createJob(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace             = "default"
		expectedJobName               = "job-gain"
		expectedImage                 = "busybox:1.28"
		expectedParallelism           = int32(2)
		expectedCompletions           = int32(4)
		expectedBackoffLimit          = int32(3)
		expectedActiveDeadlineSeconds = int64(40)
	)

	job, err := clientset.BatchV1().Jobs(expectedNamespace).Get(context.TODO(), expectedJobName, metav1.GetOptions{})

	passed := err == nil &&
		expectedParallelism == *job.Spec.Parallelism &&
		expectedCompletions == *job.Spec.Completions &&
		expectedBackoffLimit == *job.Spec.BackoffLimit &&
		expectedActiveDeadlineSeconds == *job.Spec.ActiveDeadlineSeconds

	return Result{
		TestName:   "Question 24 - Create a job job-gain with parallelism 2, completions 4, backoffLimit 3 and deadlineSeconds 40",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

func createCronjob(clientset *kubernetes.Clientset) Result {
	const (
		expectedScheduleTime  = "*/5 * * * *"
		expectedName          = "cronjob-gain"
		expectedJobName       = "job-gain"
		expectedNamespace     = "default"
		expectedImageName     = "busybox:1.28"
		expectedCommand       = "sleep 3600"
		expectedRestartPolicy = "Never"
	)

	cronjob, err := clientset.BatchV1().CronJobs(expectedNamespace).Get(context.TODO(), expectedName, metav1.GetOptions{})

	passed := err == nil &&
		expectedScheduleTime == cronjob.Spec.Schedule &&
		expectedName == cronjob.Name &&
		expectedJobName == cronjob.Spec.JobTemplate.Name &&
		expectedImageName == cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image &&
		expectedCommand == cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Command[0] &&
		expectedRestartPolicy == cronjob.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy

	return Result{
		TestName:   "Question 25 - Create a cronjob echo run a each 5 minutes with image busybox:1.28, command 'sleep 3600' and restartPolicy Never",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

func createStateFulSet(clientset *kubernetes.Clientset) Result {
	const (
		expectedNamespace = "default"
		expectedName      = "statefulset-gain"
		expectedImage     = "busybox:1.28"
		expectedCommand   = "sleep 3600"
		expectedReplicas  = int32(3)
	)

	statefulset, err := clientset.AppsV1().StatefulSets(expectedNamespace).Get(context.TODO(), expectedName, metav1.GetOptions{})

	passed := err == nil &&
		expectedName == statefulset.Name &&
		expectedImage == statefulset.Spec.Template.Spec.Containers[0].Image &&
		expectedCommand == statefulset.Spec.Template.Spec.Containers[0].Command[0] &&
		expectedReplicas == statefulset.Status.ReadyReplicas

	return Result{
		TestName:   "Question 26 - Create a statefulset statefulset-gain with image busybox:1.28, command 'sleep 3600' and replicas 3",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

// render table of results
func renderResultsTable(results []Result) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"KubeLearn - Test your knowledge of Kubernetes v0.1.8", "Result", "Difficulty"})
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

// hasCorrectIngressRule checks if the network policy has correct ingress rule this is a function
func hasCorrectIngressRule(ingressRules []v1.NetworkPolicyIngressRule) bool {
	for _, rule := range ingressRules {
		for _, from := range rule.From {
			if from.PodSelector != nil {
				selector, err := metav1.LabelSelectorAsSelector(from.PodSelector)
				if err == nil && selector.Matches(labels.Set{"tier": "frontend"}) {
					for _, port := range rule.Ports {
						if port.Port != nil && port.Port.IntVal == 6379 {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

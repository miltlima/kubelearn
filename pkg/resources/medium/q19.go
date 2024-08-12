package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func AddLivenessProbe(clientset *kubernetes.Clientset) utils.Result {
	pod, err := clientset.CoreV1().Pods("shield").Get(context.TODO(), "mark50", metav1.GetOptions{})
	passed := err == nil &&
		pod.Spec.Containers[0].LivenessProbe.InitialDelaySeconds == 5 &&
		pod.Spec.Containers[0].LivenessProbe.PeriodSeconds == 10 &&
		pod.Spec.Containers[0].LivenessProbe.HTTPGet.Path == "/" &&
		pod.Spec.Containers[0].LivenessProbe.HTTPGet.Port.IntVal == 80

	return utils.Result{
		TestName:   "Question 19 - Add a liveness probe to the pod mark50 with initial delay 5s, period 10s, HTTP GET, port 80, and path '/' in namespace shield",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

package hard

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CheckPodError(clientset *kubernetes.Clientset) utils.Result {
	pod, err := clientset.CoreV1().Pods("bandai").Get(context.TODO(), "gundamv", metav1.GetOptions{})

	passed := err == nil && pod.Spec.Containers[0].Image == "nginx:alpine"

	return utils.Result{
		TestName:   "Question 10 - Identify and fix the issue in the pod gundamv in namespace bandai",
		Passed:     passed,
		Difficulty: "Hard",
	}
}

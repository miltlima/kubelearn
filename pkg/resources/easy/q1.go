package easy

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreatePod(clientset *kubernetes.Clientset) utils.Result {
	pod, err := clientset.CoreV1().Pods("default").Get(context.TODO(), "nginx", metav1.GetOptions{})
	passed := err == nil &&
		pod.Spec.Containers[0].Image == "nginx:alpine" &&
		pod.Name == "nginx"

	return utils.Result{
		TestName:   "Question 1 - Create a pod nginx name with nginx:alpine image",
		Passed:     passed,
		Difficulty: "Easy",
	}
}

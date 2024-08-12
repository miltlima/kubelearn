package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateLabel(clientset *kubernetes.Clientset) utils.Result {
	pod, err := clientset.CoreV1().Pods("asia").Get(context.TODO(), "tshoot", metav1.GetOptions{})
	passed := err == nil && pod.Spec.Containers[0].Image == "amazon/amazon-ecs-network-sidecar:latest" && pod.ObjectMeta.Labels["country"] == "china"

	return utils.Result{
		TestName:   "Question 6 - Create a pod tshoot with label country=china with amazon/amazon-ecs-network-sidecar:latest image in namespace asia",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

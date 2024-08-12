package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateDeployment(clientset *kubernetes.Clientset) utils.Result {
	deployment, err := clientset.AppsV1().Deployments("default").Get(context.TODO(), "nginx-deployment", metav1.GetOptions{})
	passed := err == nil && deployment.Name == "nginx-deployment" && *deployment.Spec.Replicas == 4 && deployment.Spec.Template.Spec.Containers[0].Image == "nginx:alpine"

	return utils.Result{
		TestName:   "Question 2 - Create a deployment nginx-deployment with nginx:alpine image and 4 replicas",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

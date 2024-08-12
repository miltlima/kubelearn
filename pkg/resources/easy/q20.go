package easy

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateDeploymentYellow(clientset *kubernetes.Clientset) utils.Result {
	deployment, err := clientset.AppsV1().Deployments("colors").Get(context.TODO(), "yellow-deployment", metav1.GetOptions{})
	passed := err == nil && deployment.Spec.Template.Spec.Containers[0].Image == "bonovoo/node-app:1.0" && *deployment.Spec.Replicas == 2

	return utils.Result{
		TestName:   "Question 20 - Create a deployment yellow-deployment with bonovoo/node-app:1.0 image and 2 replicas in namespace colors",
		Passed:     passed,
		Difficulty: "Easy",
	}
}

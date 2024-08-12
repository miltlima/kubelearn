package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func AddServiceAccountToDeployment(clientset *kubernetes.Clientset) utils.Result {
	deploy, err := clientset.AppsV1().Deployments("default").Get(context.TODO(), "mark42", metav1.GetOptions{})
	passed := err == nil && deploy.Spec.Template.Spec.ServiceAccountName == "america-sa"

	return utils.Result{
		TestName:   "Question 15 - Add service account america-sa to the deployment mark42",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

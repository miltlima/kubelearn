package medium

import (
	"context"

	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ChangeReplicaCount(clientset *kubernetes.Clientset) utils.Result {

	deploy, err := clientset.AppsV1().Deployments("default").Get(context.TODO(), "mark42", metav1.GetOptions{})
	passed := err == nil && *deploy.Spec.Replicas == 5

	return utils.Result{
		TestName:   "Question 16 - Change the replica count of the deployment mark42 to 5",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

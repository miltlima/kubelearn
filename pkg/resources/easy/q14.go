package easy

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateServiceAccount(clientset *kubernetes.Clientset) utils.Result {
	sa, err := clientset.CoreV1().ServiceAccounts("default").Get(context.TODO(), "america-sa", metav1.GetOptions{})
	passed := err == nil && sa.Name == "america-sa"

	return utils.Result{
		TestName:   "Question 14 - Create a service account america-sa in default namespace",
		Passed:     passed,
		Difficulty: "Easy",
	}
}

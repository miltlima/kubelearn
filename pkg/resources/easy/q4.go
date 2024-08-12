package easy

import (
	"context"

	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateNamespace(clientset *kubernetes.Clientset) utils.Result {
	namespace, err := clientset.CoreV1().Namespaces().Get(context.TODO(), "europe", metav1.GetOptions{})
	passed := err == nil && namespace.Name == "europe"

	return utils.Result{
		TestName:   "Question 4 - Create a namespace europe",
		Passed:     passed,
		Difficulty: "Easy",
	}
}

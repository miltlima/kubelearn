package easy

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateSecret(clientset *kubernetes.Clientset) utils.Result {
	secret, err := clientset.CoreV1().Secrets("colors").Get(context.TODO(), "secret-colors", metav1.GetOptions{})
	passed := err == nil && string(secret.Data["color"]) == "red"

	return utils.Result{
		TestName:   "Question 12 - Create a secret secret-colors with data color=red in colors namespace",
		Passed:     passed,
		Difficulty: "Easy",
	}
}

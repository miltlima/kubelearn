package easy

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateServiceAccount(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	_, err := clientset.CoreV1().ServiceAccounts("default").Get(ctx, "america-sa", metav1.GetOptions{})

	var report utils.Report
	report.Found(err, "ServiceAccount", "america-sa", "default")
	return report.Result("Question 14 - Create a service account america-sa in default namespace", "Easy")
}

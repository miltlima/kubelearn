package easy

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateNamespace(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	_, err := clientset.CoreV1().Namespaces().Get(ctx, "europe", metav1.GetOptions{})

	var report utils.Report
	report.Found(err, "Namespace", "europe", "")
	return report.Result("Question 4 - Create a namespace europe", "Easy")
}

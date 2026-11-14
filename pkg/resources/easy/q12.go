package easy

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateSecret(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	secret, err := clientset.CoreV1().Secrets("colors").Get(ctx, "secret-colors", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Secret", "secret-colors", "colors") {
		color := string(secret.Data["color"])
		report.Expect(color == "red", "Key color is %s, expected red.", utils.Shown(color))
	}
	return report.Result("Question 12 - Create a secret secret-colors with data color=red in colors namespace", "Easy")
}

package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func AddServiceAccountToDeployment(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	deploy, err := clientset.AppsV1().Deployments("default").Get(ctx, "mark42", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Deployment", "mark42", "default") {
		serviceAccount := deploy.Spec.Template.Spec.ServiceAccountName
		report.Expect(serviceAccount == "america-sa", "Service account is %s, expected america-sa.", utils.Shown(serviceAccount))
	}
	return report.Result("Question 15 - Add service account america-sa to the deployment mark42", "Medium")
}

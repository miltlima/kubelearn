package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateDeployment(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	deployment, err := clientset.AppsV1().Deployments("default").Get(ctx, "nginx-deployment", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Deployment", "nginx-deployment", "default") {
		report.ExpectImage(deployment.Spec.Template.Spec, "nginx:alpine")
		replicas := deployment.Spec.Replicas
		report.Expect(utils.PtrEquals(replicas, 4), "Replicas is %s, expected 4.", utils.ShownPtr(replicas))
	}
	return report.Result("Question 2 - Create a deployment nginx-deployment with nginx:alpine image and 4 replicas", "Medium")
}

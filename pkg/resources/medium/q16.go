package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ChangeReplicaCount(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	deploy, err := clientset.AppsV1().Deployments("default").Get(ctx, "mark42", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Deployment", "mark42", "default") {
		replicas := deploy.Spec.Replicas
		report.Expect(utils.PtrEquals(replicas, 5), "Replicas is %s, expected 5.", utils.ShownPtr(replicas))
	}
	return report.Result("Question 16 - Change the replica count of the deployment mark42 to 5", "Medium")
}

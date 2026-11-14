package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func AddSecurityContext(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	deploy, err := clientset.AppsV1().Deployments("default").Get(ctx, "mark42", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Deployment", "mark42", "default") {
		container := utils.FirstContainer(deploy.Spec.Template.Spec)
		report.Expect(container != nil && container.SecurityContext != nil &&
			utils.PtrEquals(container.SecurityContext.AllowPrivilegeEscalation, false),
			"The first container should set allowPrivilegeEscalation to false in its security context.")
	}
	return report.Result("Question 18 - Prevent privilege escalation in the deployment mark42", "Medium")
}

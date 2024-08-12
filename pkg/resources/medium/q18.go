package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func AddSecurityContext(clientset *kubernetes.Clientset) utils.Result {
	deploy, err := clientset.AppsV1().Deployments("default").Get(context.TODO(), "mark42", metav1.GetOptions{})
	passed := err == nil && deploy.Spec.Template.Spec.Containers != nil && len(deploy.Spec.Template.Spec.Containers) > 0 &&
		deploy.Spec.Template.Spec.Containers[0].SecurityContext != nil &&
		*deploy.Spec.Template.Spec.Containers[0].SecurityContext.AllowPrivilegeEscalation == false

	return utils.Result{
		TestName:   "Question 18 - Prevent privilege escalation in the deployment mark42",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

package hard

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateRoleOne(clientset *kubernetes.Clientset) utils.Result {
	role, err := clientset.RbacV1().Roles("fruits").Get(context.TODO(), "apple-one", metav1.GetOptions{})

	expectedVerbs := []string{"get", "list", "watch"}
	passed := err == nil && len(role.Rules) > 0 && len(role.Rules[0].Resources) > 0 &&
		role.Rules[0].Resources[0] == "pods"

	if passed {
		for _, verb := range expectedVerbs {
			if !utils.Contains(role.Rules[0].Verbs, verb) {
				passed = false
				break
			}
		}
	}

	return utils.Result{
		TestName:   "Question 23 - Create a role apple-one with verbs get, list, watch in namespace fruits",
		Passed:     passed,
		Difficulty: "Hard",
	}
}

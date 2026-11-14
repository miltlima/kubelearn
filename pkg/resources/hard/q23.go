package hard

import (
	"context"
	"strings"

	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateRoleOne(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	role, err := clientset.RbacV1().Roles("fruits").Get(ctx, "apple-one", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Role", "apple-one", "fruits") &&
		report.Expect(len(role.Rules) > 0, "The role has no rules.") {
		rule := role.Rules[0]
		report.Expect(len(rule.Resources) > 0 && rule.Resources[0] == "pods", "The first rule should apply to the pods resource.")

		var missing []string
		for _, verb := range []string{"get", "list", "watch"} {
			if !utils.Contains(rule.Verbs, verb) {
				missing = append(missing, verb)
			}
		}
		report.Expect(len(missing) == 0, "Missing verbs: %s.", strings.Join(missing, ", "))
	}
	return report.Result("Question 23 - Create a role apple-one with verbs get, list, watch in namespace fruits", "Hard")
}

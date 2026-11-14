package hard

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateIngressYellow(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	ingress, err := clientset.NetworkingV1().Ingresses("colors").Get(ctx, "ingress-colors", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Ingress", "ingress-colors", "colors") &&
		report.Expect(len(ingress.Spec.Rules) > 0, "The ingress has no rules.") {
		rule := ingress.Spec.Rules[0]
		report.Expect(rule.Host == "yellow.com", "Host is %s, expected yellow.com.", utils.Shown(rule.Host))
		if report.Expect(rule.HTTP != nil && len(rule.HTTP.Paths) > 0, "The first rule has no HTTP paths.") {
			path := rule.HTTP.Paths[0]
			report.Expect(path.Path == "/yellow", "Path is %s, expected /yellow.", utils.Shown(path.Path))
			report.Expect(path.Backend.Service != nil && path.Backend.Service.Name == "yellow-service",
				"The path should send traffic to service yellow-service.")
		}
	}
	return report.Result("Question 22 - Create an ingress ingress-colors with host yellow.com, path /yellow, and service yellow-service in namespace colors", "Hard")
}

package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func AddLivenessProbe(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	pod, err := clientset.CoreV1().Pods("shield").Get(ctx, "mark50", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Pod", "mark50", "shield") {
		container := utils.FirstContainer(pod.Spec)
		if report.Expect(container != nil && container.LivenessProbe != nil, "The first container has no liveness probe.") {
			probe := container.LivenessProbe
			report.Expect(probe.InitialDelaySeconds == 5, "Initial delay is %ds, expected 5s.", probe.InitialDelaySeconds)
			report.Expect(probe.PeriodSeconds == 10, "Period is %ds, expected 10s.", probe.PeriodSeconds)
			report.Expect(probe.HTTPGet != nil && probe.HTTPGet.Path == "/" && probe.HTTPGet.Port.IntVal == 80,
				"The probe should be an HTTP GET on path / and port 80.")
		}
	}
	return report.Result("Question 19 - Add a liveness probe to the pod mark50 with initial delay 5s, period 10s, HTTP GET, port 80, and path '/' in namespace shield", "Medium")
}

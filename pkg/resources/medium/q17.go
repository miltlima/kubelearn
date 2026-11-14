package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateHpa(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	hpa, err := clientset.AutoscalingV2().HorizontalPodAutoscalers("default").Get(ctx, "mark43", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "HorizontalPodAutoscaler", "mark43", "default") {
		spec := hpa.Spec
		report.Expect(spec.ScaleTargetRef.Name == "mark43", "Scale target is %s, expected mark43.", utils.Shown(spec.ScaleTargetRef.Name))
		report.Expect(utils.PtrEquals(spec.MinReplicas, 2), "Min replicas is %s, expected 2.", utils.ShownPtr(spec.MinReplicas))
		report.Expect(spec.MaxReplicas == 8, "Max replicas is %d, expected 8.", spec.MaxReplicas)
		report.Expect(len(spec.Metrics) > 0 && spec.Metrics[0].Resource != nil &&
			utils.PtrEquals(spec.Metrics[0].Resource.Target.AverageUtilization, 80),
			"The first metric should target 80%% average utilization.")
	}
	return report.Result("Question 17 - Create a horizontal pod autoscaler hpa-mark43 for deployment mark43 with CPU utilization 80%, min replicas 2 and max replicas 8", "Medium")
}

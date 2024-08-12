package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateHpa(clientset *kubernetes.Clientset) utils.Result {
	hpa, err := clientset.AutoscalingV2().HorizontalPodAutoscalers("default").Get(context.TODO(), "mark43", metav1.GetOptions{})
	passed := err == nil && hpa.Spec.ScaleTargetRef.Name == "mark43" && *hpa.Spec.MinReplicas == 2 && hpa.Spec.MaxReplicas == 8 && *hpa.Spec.Metrics[0].Resource.Target.AverageUtilization == 80

	return utils.Result{
		TestName:   "Question 17 - Create a horizontal pod autoscaler hpa-mark43 for deployment mark43 with CPU utilization 80%, min replicas 2 and max replicas 8",
		Passed:     passed,
		Difficulty: "Medium",
	}
}

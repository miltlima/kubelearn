package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateConfigMap(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	configMap, err := clientset.CoreV1().ConfigMaps("default").Get(ctx, "europe-configmap", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "ConfigMap", "europe-configmap", "default") {
		france := configMap.Data["France"]
		report.Expect(france == "Paris", "Key France is %s, expected Paris.", utils.Shown(france))
	}
	return report.Result("Question 5 - Create a configmap europe-configmap with data France=Paris", "Medium")
}

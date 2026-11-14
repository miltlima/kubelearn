package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreatePersistentVolumeClaim(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	pvc, err := clientset.CoreV1().PersistentVolumeClaims("default").Get(ctx, "unicorn-pvc", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "PersistentVolumeClaim", "unicorn-pvc", "default") {
		requested := pvc.Spec.Resources.Requests.Storage().String()
		report.Expect(requested == "400Mi", "Requested storage is %s, expected 400Mi.", requested)
		report.Expect(len(pvc.Spec.AccessModes) > 0 && pvc.Spec.AccessModes[0] == "ReadWriteMany",
			"The first access mode should be ReadWriteMany.")
	}
	return report.Result("Question 8 - Create a persistent volume claim unicorn-pvc with capacity 400Mi and access mode ReadWriteMany", "Medium")
}

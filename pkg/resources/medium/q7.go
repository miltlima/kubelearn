package medium

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreatePersistentVolume(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	pv, err := clientset.CoreV1().PersistentVolumes().Get(ctx, "unicorn-pv", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "PersistentVolume", "unicorn-pv", "") {
		capacity := pv.Spec.Capacity.Storage().String()
		report.Expect(capacity == "1Gi", "Capacity is %s, expected 1Gi.", capacity)
		report.Expect(len(pv.Spec.AccessModes) > 0 && pv.Spec.AccessModes[0] == "ReadWriteMany",
			"The first access mode should be ReadWriteMany.")
		report.Expect(pv.Spec.HostPath != nil && pv.Spec.HostPath.Path == "/tmp/data",
			"The volume should use host path /tmp/data.")
	}
	return report.Result("Question 7 - Create a persistent volume unicorn-pv with capacity 1Gi, access mode ReadWriteMany, and host path /tmp/data", "Medium")
}

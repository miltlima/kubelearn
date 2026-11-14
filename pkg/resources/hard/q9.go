package hard

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreatePodVolumeClaim(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	pod, err := clientset.CoreV1().Pods("public").Get(ctx, "webserver", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Pod", "webserver", "public") {
		container := report.ExpectImage(pod.Spec, "nginx:alpine")
		if report.Expect(len(pod.Spec.Volumes) > 0, "The pod has no volumes.") {
			volume := pod.Spec.Volumes[0]
			report.Expect(volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == "unicorn-pvc",
				"The first volume should use persistent volume claim unicorn-pvc.")
			report.Expect(volume.Name == "unicorn-pv", "The first volume is named %s, expected unicorn-pv.", volume.Name)
		}
		report.Expect(container != nil && len(container.VolumeMounts) > 0 &&
			container.VolumeMounts[0].MountPath == "/usr/share/nginx/html",
			"The container's first volume mount should be /usr/share/nginx/html.")
	}
	return report.Result("Question 9 - Create a pod webserver in public namespace with nginx:alpine image, volume mount /usr/share/nginx/html, and a persistent volume claim unicorn-pvc", "Hard")
}

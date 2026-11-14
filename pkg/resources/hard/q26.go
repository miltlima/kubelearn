package hard

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateStatefulSet(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	statefulset, err := clientset.AppsV1().StatefulSets("default").Get(ctx, "statefulset-gain", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "StatefulSet", "statefulset-gain", "default") {
		container := report.ExpectImage(statefulset.Spec.Template.Spec, "busybox:1.28")
		report.Expect(utils.RunsCommand(container, "sleep 3600"), "The container should run the command sleep 3600.")
		ready := statefulset.Status.ReadyReplicas
		report.Expect(ready == 3, "%d of 3 replicas are ready.", ready)
	}
	return report.Result("Question 26 - Create a statefulset statefulset-gain with image busybox:1.28, command 'sleep 3600', and 3 replicas", "Hard")
}

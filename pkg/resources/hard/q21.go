package hard

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateServiceForYellow(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	service, err := clientset.CoreV1().Services("colors").Get(ctx, "yellow-service", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Service", "yellow-service", "colors") {
		ports := service.Spec.Ports
		report.Expect(len(ports) > 0 && ports[0].Port == 80 && ports[0].TargetPort.IntVal == 3000,
			"The service's first port should be 80 with target port 3000.")
		app := service.Spec.Selector["app"]
		report.Expect(app == "yellow-deployment", "Selector app is %s, expected yellow-deployment.", utils.Shown(app))
	}
	return report.Result("Question 21 - Create a service yellow-service for the deployment yellow-deployment in namespace colors with port 80 and target port 3000", "Hard")
}

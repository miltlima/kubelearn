package resources_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	"kubelearn/pkg/resources"
	"kubelearn/pkg/resources/medium"
	"kubelearn/pkg/utils"
)

type check struct {
	name string
	fn   resources.Check
}

// allChecks names every registered check after its question number.
var allChecks = func() []check {
	checks := make([]check, 0, len(resources.Checks))
	for i, fn := range resources.Checks {
		checks = append(checks, check{fmt.Sprintf("q%d", i+1), fn})
	}
	return checks
}()

// runCheck runs one checker and converts a panic into a test failure.
func runCheck(t *testing.T, c check, cs kubernetes.Interface) (passed bool) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s panicked: %v", c.name, r)
		}
	}()
	return c.fn(context.Background(), cs).Passed
}

func ptr[T any](v T) *T { return &v }

func meta(namespace, name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{Namespace: namespace, Name: name}
}

func podSpec(image string) corev1.PodSpec {
	return corev1.PodSpec{Containers: []corev1.Container{{Name: "main", Image: image}}}
}

func deployment(namespace, name string, replicas int32, spec corev1.PodSpec) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: meta(namespace, name),
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr(replicas),
			Template: corev1.PodTemplateSpec{Spec: spec},
		},
	}
}

// solvedCluster returns objects that satisfy every question.
func solvedCluster() []runtime.Object {
	mark42 := podSpec("nginx:alpine")
	mark42.ServiceAccountName = "america-sa"
	mark42.Containers[0].SecurityContext = &corev1.SecurityContext{AllowPrivilegeEscalation: ptr(false)}

	webserver := podSpec("nginx:alpine")
	webserver.Containers[0].VolumeMounts = []corev1.VolumeMount{{Name: "unicorn-pv", MountPath: "/usr/share/nginx/html"}}
	webserver.Volumes = []corev1.Volume{{
		Name: "unicorn-pv",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "unicorn-pvc"},
		},
	}}

	purple := podSpec("redis:alpine")
	purple.Volumes = []corev1.Volume{{
		Name:         "secret",
		VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: "secret-purple"}},
	}}

	mark50 := podSpec("nginx:alpine")
	mark50.Containers[0].LivenessProbe = &corev1.Probe{
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{Path: "/", Port: intstr.FromInt(80)},
		},
	}

	sleeper := podSpec("busybox:1.28")
	sleeper.Containers[0].Command = []string{"sleep 3600"}
	cronSleeper := *sleeper.DeepCopy()
	cronSleeper.RestartPolicy = corev1.RestartPolicyNever

	pathType := networkingv1.PathTypePrefix

	return []runtime.Object{
		&corev1.Pod{ObjectMeta: meta("default", "nginx"), Spec: podSpec("nginx:alpine")},
		deployment("default", "nginx-deployment", 4, podSpec("nginx:alpine")),
		deployment("latam", "redis", 1, podSpec("redis:alpine")),
		&corev1.Service{
			ObjectMeta: meta("latam", "redis-service"),
			Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 6379}}},
		},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "europe"}},
		&corev1.ConfigMap{ObjectMeta: meta("default", "europe-configmap"), Data: map[string]string{"France": "Paris"}},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Namespace: "asia", Name: "tshoot", Labels: map[string]string{"country": "china"}},
			Spec:       podSpec("amazon/amazon-ecs-network-sidecar:latest"),
		},
		&corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{Name: "unicorn-pv"},
			Spec: corev1.PersistentVolumeSpec{
				Capacity:    corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("1Gi")},
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
				PersistentVolumeSource: corev1.PersistentVolumeSource{
					HostPath: &corev1.HostPathVolumeSource{Path: "/tmp/data"},
				},
			},
		},
		&corev1.PersistentVolumeClaim{
			ObjectMeta: meta("default", "unicorn-pvc"),
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("400Mi")},
				},
			},
		},
		&corev1.Pod{ObjectMeta: meta("public", "webserver"), Spec: webserver},
		&corev1.Pod{ObjectMeta: meta("bandai", "gundamv"), Spec: podSpec("nginx:alpine")},
		&networkingv1.NetworkPolicy{
			ObjectMeta: meta("colors", "allow-policy-colors"),
			Spec: networkingv1.NetworkPolicySpec{
				Ingress: []networkingv1.NetworkPolicyIngressRule{{
					From: []networkingv1.NetworkPolicyPeer{{
						PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"tier": "frontend"}},
					}},
					Ports: []networkingv1.NetworkPolicyPort{{Port: ptr(intstr.FromInt(6379))}},
				}},
			},
		},
		&corev1.Secret{ObjectMeta: meta("colors", "secret-colors"), Data: map[string][]byte{"color": []byte("red")}},
		&corev1.Pod{ObjectMeta: meta("colors", "purple"), Spec: purple},
		&corev1.Secret{ObjectMeta: meta("colors", "secret-purple"), Data: map[string][]byte{"singer": []byte("prince")}},
		&corev1.ServiceAccount{ObjectMeta: meta("default", "america-sa")},
		deployment("default", "mark42", 5, mark42),
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: meta("default", "mark43"),
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{Kind: "Deployment", Name: "mark43"},
				MinReplicas:    ptr(int32(2)),
				MaxReplicas:    8,
				Metrics: []autoscalingv2.MetricSpec{{
					Type: autoscalingv2.ResourceMetricSourceType,
					Resource: &autoscalingv2.ResourceMetricSource{
						Name: corev1.ResourceCPU,
						Target: autoscalingv2.MetricTarget{
							Type:               autoscalingv2.UtilizationMetricType,
							AverageUtilization: ptr(int32(80)),
						},
					},
				}},
			},
		},
		&corev1.Pod{ObjectMeta: meta("shield", "mark50"), Spec: mark50},
		deployment("colors", "yellow-deployment", 2, podSpec("bonovoo/node-app:1.0")),
		&corev1.Service{
			ObjectMeta: meta("colors", "yellow-service"),
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"app": "yellow-deployment"},
				Ports:    []corev1.ServicePort{{Port: 80, TargetPort: intstr.FromInt(3000)}},
			},
		},
		&networkingv1.Ingress{
			ObjectMeta: meta("colors", "ingress-colors"),
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{{
					Host: "yellow.com",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{{
								Path:     "/yellow",
								PathType: &pathType,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{Name: "yellow-service"},
								},
							}},
						},
					},
				}},
			},
		},
		&rbacv1.Role{
			ObjectMeta: meta("fruits", "apple-one"),
			Rules:      []rbacv1.PolicyRule{{Resources: []string{"pods"}, Verbs: []string{"get", "list", "watch"}}},
		},
		&batchv1.Job{
			ObjectMeta: meta("default", "job-gain"),
			Spec: batchv1.JobSpec{
				Parallelism:           ptr(int32(2)),
				Completions:           ptr(int32(4)),
				BackoffLimit:          ptr(int32(3)),
				ActiveDeadlineSeconds: ptr(int64(40)),
			},
		},
		&batchv1.CronJob{
			ObjectMeta: meta("default", "cronjob-gain"),
			Spec: batchv1.CronJobSpec{
				Schedule: "*/5 * * * *",
				JobTemplate: batchv1.JobTemplateSpec{
					Spec: batchv1.JobSpec{Template: corev1.PodTemplateSpec{Spec: cronSleeper}},
				},
			},
		},
		&appsv1.StatefulSet{
			ObjectMeta: meta("default", "statefulset-gain"),
			Spec:       appsv1.StatefulSetSpec{Template: corev1.PodTemplateSpec{Spec: sleeper}},
			Status:     appsv1.StatefulSetStatus{ReadyReplicas: 3},
		},
	}
}

// stripped returns the solved cluster with every object reduced according to
// level, keeping only names so the checkers still find the objects:
//
//	0: empty specs (no containers, volumes, ports, rules, replicas...)
//	1: one bare container / volume / port / rule with no nested fields set
func stripped(level int) []runtime.Object {
	bare := corev1.PodSpec{}
	if level == 1 {
		bare = corev1.PodSpec{
			Containers: []corev1.Container{{Name: "main", SecurityContext: &corev1.SecurityContext{}}},
			Volumes:    []corev1.Volume{{Name: "data"}},
		}
	}
	template := corev1.PodTemplateSpec{Spec: bare}

	var objects []runtime.Object
	for _, obj := range solvedCluster() {
		switch o := obj.(type) {
		case *corev1.Pod:
			o.Spec = *bare.DeepCopy()
		case *appsv1.Deployment:
			o.Spec = appsv1.DeploymentSpec{Template: *template.DeepCopy()}
		case *appsv1.StatefulSet:
			o.Spec = appsv1.StatefulSetSpec{Template: *template.DeepCopy()}
		case *batchv1.Job:
			o.Spec = batchv1.JobSpec{}
		case *batchv1.CronJob:
			o.Spec = batchv1.CronJobSpec{
				Schedule:    "*/5 * * * *",
				JobTemplate: batchv1.JobTemplateSpec{Spec: batchv1.JobSpec{Template: *template.DeepCopy()}},
			}
		case *corev1.Service:
			o.Spec = corev1.ServiceSpec{}
			if level == 1 {
				o.Spec.Ports = []corev1.ServicePort{{}}
			}
		case *corev1.PersistentVolume:
			o.Spec = corev1.PersistentVolumeSpec{}
			if level == 1 {
				o.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany}
				o.Spec.Capacity = corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("1Gi")}
			}
		case *corev1.PersistentVolumeClaim:
			o.Spec = corev1.PersistentVolumeClaimSpec{}
		case *networkingv1.NetworkPolicy:
			o.Spec = networkingv1.NetworkPolicySpec{}
			if level == 1 {
				o.Spec.Ingress = []networkingv1.NetworkPolicyIngressRule{{
					From:  []networkingv1.NetworkPolicyPeer{{}},
					Ports: []networkingv1.NetworkPolicyPort{{}},
				}}
			}
		case *networkingv1.Ingress:
			o.Spec = networkingv1.IngressSpec{}
			if level == 1 {
				o.Spec.Rules = []networkingv1.IngressRule{{Host: "yellow.com"}}
			}
		case *autoscalingv2.HorizontalPodAutoscaler:
			o.Spec = autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{Name: "mark43"},
				MaxReplicas:    8,
			}
			if level == 1 {
				o.Spec.MinReplicas = ptr(int32(2))
				o.Spec.Metrics = []autoscalingv2.MetricSpec{{Type: autoscalingv2.PodsMetricSourceType}}
			}
		case *rbacv1.Role:
			o.Rules = nil
			if level == 1 {
				o.Rules = []rbacv1.PolicyRule{{}}
			}
		}
		objects = append(objects, obj)
	}
	return objects
}

func TestSolvedClusterPassesEveryQuestion(t *testing.T) {
	cs := fake.NewSimpleClientset(solvedCluster()...)
	for _, c := range allChecks {
		if !runCheck(t, c, cs) {
			t.Errorf("%s: expected Passed=true on a solved cluster", c.name)
		}
	}
}

func TestEmptyClusterFailsEveryQuestionWithoutPanic(t *testing.T) {
	cs := fake.NewSimpleClientset()
	for _, c := range allChecks {
		if runCheck(t, c, cs) {
			t.Errorf("%s: expected Passed=false on an empty cluster", c.name)
		}
	}
}

// Objects that exist but are incomplete must fail the question, never panic.
func TestIncompleteObjectsDoNotPanic(t *testing.T) {
	for level := 0; level <= 1; level++ {
		t.Run(fmt.Sprintf("level%d", level), func(t *testing.T) {
			cs := fake.NewSimpleClientset(stripped(level)...)
			for _, c := range allChecks {
				runCheck(t, c, cs)
			}
		})
	}
}

// Questions 3 and 13 read two objects; the second one existing must not hide
// that the first one is missing.
func TestSecondObjectAloneDoesNotPass(t *testing.T) {
	var objects []runtime.Object
	for _, obj := range solvedCluster() {
		switch o := obj.(type) {
		case *corev1.Service:
			if o.Name == "redis-service" {
				objects = append(objects, o)
			}
		case *corev1.Secret:
			if o.Name == "secret-purple" {
				objects = append(objects, o)
			}
		}
	}
	cs := fake.NewSimpleClientset(objects...)
	for _, c := range allChecks {
		if c.name != "q3" && c.name != "q13" {
			continue
		}
		if runCheck(t, c, cs) {
			t.Errorf("%s: expected Passed=false when only the second object exists", c.name)
		}
	}
}

// Regression: pod mark50 existing in namespace shield without a liveness probe
// (the state before the question is solved) used to crash /questions and
// /finish.
func TestLivenessProbeQuestionOnFreshLab(t *testing.T) {
	fresh := &corev1.Pod{ObjectMeta: meta("shield", "mark50"), Spec: podSpec("nginx:alpine")}
	cs := fake.NewSimpleClientset(fresh)
	c := check{"q19", medium.AddLivenessProbe}
	if runCheck(t, c, cs) {
		t.Errorf("q19: expected Passed=false for a pod without liveness probe")
	}
}

// The registry position defines the question number shown to the user.
func TestRegistryOrderMatchesQuestionNumbers(t *testing.T) {
	results := evaluate(t, fake.NewSimpleClientset())
	if len(results) != len(resources.Checks) {
		t.Fatalf("Evaluate returned %d results for %d checks", len(results), len(resources.Checks))
	}
	for i, result := range results {
		prefix := fmt.Sprintf("Question %d - ", i+1)
		if !strings.HasPrefix(result.TestName, prefix) {
			t.Errorf("position %d holds %q, want a name starting with %q", i+1, result.TestName, prefix)
		}
	}
}

func TestScore(t *testing.T) {
	if got := resources.Score(nil); got != 0 {
		t.Errorf("Score(nil) = %v, want 0", got)
	}
	if got := resources.Score(evaluate(t, fake.NewSimpleClientset())); got != 0 {
		t.Errorf("empty cluster score = %v, want 0", got)
	}
	if got := resources.Score(evaluate(t, fake.NewSimpleClientset(solvedCluster()...))); got != 100 {
		t.Errorf("solved cluster score = %v, want 100", got)
	}
	half := []utils.Result{{Passed: true}, {Passed: false}}
	if got := resources.Score(half); got != 50 {
		t.Errorf("half passed score = %v, want 50", got)
	}
}

func evaluate(t *testing.T, cs kubernetes.Interface) []utils.Result {
	t.Helper()
	results, err := resources.Evaluate(context.Background(), cs)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	return results
}

// A cancelled or expired context must surface as an error, not as a list of
// failed questions.
func TestEvaluateReportsContextErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results, err := resources.Evaluate(ctx, fake.NewSimpleClientset(solvedCluster()...))
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
	if results != nil {
		t.Errorf("got %d results from a cancelled evaluation, want none", len(results))
	}
}

// result runs the named question against a cluster holding objects.
func result(t *testing.T, name string, objects ...runtime.Object) utils.Result {
	t.Helper()
	cs := fake.NewSimpleClientset(objects...)
	for _, c := range allChecks {
		if c.name == name {
			return c.fn(context.Background(), cs)
		}
	}
	t.Fatalf("no check named %s", name)
	return utils.Result{}
}

// Every unsolved question must say why; a solved one must not.
func TestFeedbackIsPresentExactlyWhenFailing(t *testing.T) {
	for _, r := range evaluate(t, fake.NewSimpleClientset()) {
		if r.Feedback == "" {
			t.Errorf("%q failed on an empty cluster without feedback", r.TestName)
		}
	}
	for level := 0; level <= 1; level++ {
		for _, r := range evaluate(t, fake.NewSimpleClientset(stripped(level)...)) {
			if r.Passed != (r.Feedback == "") {
				t.Errorf("level %d: %q has Passed=%v with feedback %q", level, r.TestName, r.Passed, r.Feedback)
			}
		}
	}
	for _, r := range evaluate(t, fake.NewSimpleClientset(solvedCluster()...)) {
		if r.Feedback != "" {
			t.Errorf("%q passed but carries feedback %q", r.TestName, r.Feedback)
		}
	}
}

func TestFeedbackMessages(t *testing.T) {
	wrongImage := &corev1.Pod{ObjectMeta: meta("default", "nginx"), Spec: podSpec("nginx:latest")}
	noProbe := &corev1.Pod{ObjectMeta: meta("shield", "mark50"), Spec: podSpec("nginx:alpine")}
	oneReplica := deployment("default", "nginx-deployment", 1, podSpec("nginx:alpine"))
	readOnlyRole := &rbacv1.Role{
		ObjectMeta: meta("fruits", "apple-one"),
		Rules:      []rbacv1.PolicyRule{{Resources: []string{"pods"}, Verbs: []string{"get"}}},
	}

	cases := []struct {
		question string
		objects  []runtime.Object
		want     string
	}{
		{"q1", nil, "Pod nginx not found in namespace default."},
		{"q4", nil, "Namespace europe not found."},
		{"q1", []runtime.Object{wrongImage}, "Container image is nginx:latest, expected nginx:alpine."},
		{"q2", []runtime.Object{oneReplica}, "Replicas is 1, expected 4."},
		{"q19", []runtime.Object{noProbe}, "The first container has no liveness probe."},
		{"q23", []runtime.Object{readOnlyRole}, "Missing verbs: list, watch."},
		{"q3", nil, "Deployment redis not found in namespace latam. Service redis-service not found in namespace latam."},
	}
	for _, tc := range cases {
		if got := result(t, tc.question, tc.objects...).Feedback; got != tc.want {
			t.Errorf("%s: feedback %q, want %q", tc.question, got, tc.want)
		}
	}
}

// The troubleshooting question must not give the answer away.
func TestTroubleshootingFeedbackDoesNotRevealTheFix(t *testing.T) {
	broken := &corev1.Pod{ObjectMeta: meta("bandai", "gundamv"), Spec: podSpec("nginxxx:alpine")}
	feedback := result(t, "q10", broken).Feedback
	if feedback == "" || strings.Contains(feedback, "nginx") || strings.Contains(feedback, "image") {
		t.Errorf("feedback %q should hint without naming the image", feedback)
	}
}

// A request that fails for a reason other than "not found" is a cluster
// problem and must be reported as such.
func TestFeedbackReportsClusterErrors(t *testing.T) {
	cs := fake.NewSimpleClientset()
	cs.PrependReactor("get", "pods", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("connection refused")
	})
	r := resources.Checks[0](context.Background(), cs)
	if r.Passed || !strings.Contains(r.Feedback, "Could not read pod nginx from the cluster") ||
		!strings.Contains(r.Feedback, "connection refused") {
		t.Errorf("unexpected result for an unreachable cluster: %+v", r)
	}
}

// "sleep 3600" must be accepted however the command is split; the single
// string ["sleep 3600"] cannot even start in a real container.
func TestSleepCommandIsAcceptedInAnyForm(t *testing.T) {
	forms := [][2][]string{
		{{"sleep", "3600"}, nil},
		{{"sh", "-c", "sleep 3600"}, nil},
		{{"sleep"}, {"3600"}},
		{nil, {"sleep", "3600"}},
		{{"sleep 3600"}, nil},
	}
	for _, form := range forms {
		spec := podSpec("busybox:1.28")
		spec.Containers[0].Command, spec.Containers[0].Args = form[0], form[1]
		sts := &appsv1.StatefulSet{
			ObjectMeta: meta("default", "statefulset-gain"),
			Spec:       appsv1.StatefulSetSpec{Template: corev1.PodTemplateSpec{Spec: spec}},
			Status:     appsv1.StatefulSetStatus{ReadyReplicas: 3},
		}
		if r := result(t, "q26", sts); !r.Passed {
			t.Errorf("command %v args %v rejected: %s", form[0], form[1], r.Feedback)
		}
	}

	wrong := podSpec("busybox:1.28")
	wrong.Containers[0].Command = []string{"sleep", "60"}
	sts := &appsv1.StatefulSet{
		ObjectMeta: meta("default", "statefulset-gain"),
		Spec:       appsv1.StatefulSetSpec{Template: corev1.PodTemplateSpec{Spec: wrong}},
		Status:     appsv1.StatefulSetStatus{ReadyReplicas: 3},
	}
	if r := result(t, "q26", sts); r.Passed {
		t.Errorf("command sleep 60 was accepted")
	}
}

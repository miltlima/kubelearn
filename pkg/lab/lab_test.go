package lab

import (
	"context"
	"errors"
	"strings"
	"testing"
	"testing/fstest"
)

// recorder is a fake command runner: it records every command and answers
// from a list of rules.
type recorder struct {
	calls    []string
	stdin    map[string]string
	clusters string            // output of "kind get clusters"
	fail     map[string]string // command prefix -> error message
}

func (r *recorder) run(_ context.Context, stdin, name string, args ...string) (string, error) {
	command := name + " " + strings.Join(args, " ")
	r.calls = append(r.calls, command)
	if r.stdin == nil {
		r.stdin = map[string]string{}
	}
	r.stdin[command] += stdin
	for prefix, message := range r.fail {
		if strings.HasPrefix(command, prefix) {
			return message, errors.New("exit status 1")
		}
	}
	if command == "kind get clusters" {
		return r.clusters, nil
	}
	return "", nil
}

func found(string) (string, error) { return "/usr/bin/tool", nil }

func testLab(r *recorder) (*Lab, *strings.Builder) {
	out := &strings.Builder{}
	manifests := fstest.MapFS{
		"klearn-0001.yaml":          {Data: []byte("kind: Namespace # one")},
		"klearn-0000.yaml":          {Data: []byte("kind: Namespace # zero")},
		"notes.txt":                 {Data: []byte("not a manifest")},
		"addons/gateway-class.yaml": {Data: []byte("kind: GatewayClass")},
	}
	return &Lab{Manifests: manifests, Out: out, Run: r.run, LookPath: found}, out
}

func position(calls []string, prefix string) int {
	for i, call := range calls {
		if strings.HasPrefix(call, prefix) {
			return i
		}
	}
	return -1
}

func TestUpCreatesTheClusterAndAppliesEverythingInOrder(t *testing.T) {
	r := &recorder{clusters: "other\n"}
	l, _ := testLab(r)

	if err := l.Up(context.Background()); err != nil {
		t.Fatal(err)
	}

	steps := []string{
		"kind get clusters",
		"kind create cluster --name kubelearn --image " + NodeImage,
		"kubectl config use-context kind-kubelearn",
		"kubectl --context kind-kubelearn apply -f -",
		"kubectl --context kind-kubelearn apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/" + GatewayAPIVersion,
		"kubectl --context kind-kubelearn wait --for=condition=Established",
	}
	last := -1
	for _, step := range steps {
		at := position(r.calls, step)
		if at <= last {
			t.Fatalf("step %q missing or out of order in:\n%s", step, strings.Join(r.calls, "\n"))
		}
		last = at
	}

	// Lab manifests in name order, then the addons; nothing else.
	applied := r.stdin["kubectl --context kind-kubelearn apply -f -"]
	if applied != "kind: Namespace # zerokind: Namespace # onekind: GatewayClass" {
		t.Errorf("manifests applied: %q", applied)
	}
	if !strings.Contains(r.stdin[r.calls[position(r.calls, "kind create cluster")]], "role: worker") {
		t.Errorf("the kind configuration was not passed to kind create")
	}
}

func TestUpNeverTouchesAnotherContext(t *testing.T) {
	r := &recorder{}
	l, _ := testLab(r)
	if err := l.Up(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, call := range r.calls {
		if strings.HasPrefix(call, "kubectl ") && !strings.Contains(call, "kind-kubelearn") {
			t.Errorf("kubectl command without the lab context: %s", call)
		}
	}
}

func TestUpReusesAnExistingCluster(t *testing.T) {
	r := &recorder{clusters: "kubelearn\n"}
	l, out := testLab(r)

	if err := l.Up(context.Background()); err != nil {
		t.Fatal(err)
	}
	if position(r.calls, "kind create cluster") != -1 {
		t.Errorf("created a cluster that already exists")
	}
	if position(r.calls, "kubectl --context kind-kubelearn apply -f -") == -1 {
		t.Errorf("did not apply the lab resources to the existing cluster")
	}
	if !strings.Contains(out.String(), "already exists") {
		t.Errorf("output does not say the cluster exists: %q", out.String())
	}
}

func TestUpContinuesWithoutTheGatewayAPI(t *testing.T) {
	r := &recorder{fail: map[string]string{"kubectl --context kind-kubelearn apply -f https://": "no route to host"}}
	l, out := testLab(r)

	if err := l.Up(context.Background()); err != nil {
		t.Fatalf("a failed Gateway API download must not fail the lab: %v", err)
	}
	if !strings.Contains(out.String(), "questions 51 to 54") || !strings.Contains(out.String(), "no route to host") {
		t.Errorf("warning missing from output: %q", out.String())
	}
}

func TestUpFailsWhenTheClusterCannotBeCreated(t *testing.T) {
	r := &recorder{fail: map[string]string{"kind create cluster": "docker daemon is not running"}}
	l, _ := testLab(r)

	err := l.Up(context.Background())
	if err == nil || !strings.Contains(err.Error(), "docker daemon is not running") {
		t.Fatalf("err = %v, want the command output", err)
	}
	if position(r.calls, "kubectl") != -1 {
		t.Errorf("ran kubectl after the cluster creation failed")
	}
}

func TestUpNamesTheMissingTool(t *testing.T) {
	r := &recorder{}
	l, _ := testLab(r)
	l.LookPath = func(tool string) (string, error) {
		if tool == "kind" {
			return "", errors.New("not found")
		}
		return "/usr/bin/" + tool, nil
	}

	err := l.Up(context.Background())
	if err == nil || !strings.Contains(err.Error(), "kind is not installed") {
		t.Fatalf("err = %v", err)
	}
	if len(r.calls) != 0 {
		t.Errorf("ran commands without kind: %v", r.calls)
	}
}

func TestDownDeletesOnlyTheLabCluster(t *testing.T) {
	r := &recorder{}
	l, _ := testLab(r)

	if err := l.Down(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(r.calls) != 1 || r.calls[0] != "kind delete cluster --name kubelearn" {
		t.Errorf("calls = %v", r.calls)
	}
}

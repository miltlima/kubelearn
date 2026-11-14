// Package lab creates and removes the local practice cluster.
//
// It drives the kind and kubectl command-line tools, which the learner needs
// anyway to solve the questions.
package lab

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os/exec"
	"sort"
	"strings"
)

const (
	// ClusterName is the name of the kind cluster.
	ClusterName = "kubelearn"
	// Context is the kubectl context kind creates for the cluster.
	Context = "kind-" + ClusterName
	// NodeImage is the Kubernetes version the questions are written for.
	NodeImage = "kindest/node:v1.27.1"
	// GatewayAPIVersion is the Gateway API release installed for the
	// networking questions.
	GatewayAPIVersion = "v1.2.1"
)

// One control plane and two workers.
const kindConfig = `kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
`

var gatewayCRDs = []string{
	"crd/gatewayclasses.gateway.networking.k8s.io",
	"crd/gateways.gateway.networking.k8s.io",
	"crd/httproutes.gateway.networking.k8s.io",
}

// Runner runs a command, feeding it stdin when not empty, and returns
// everything it printed.
type Runner func(ctx context.Context, stdin, name string, args ...string) (output string, err error)

// Lab manages the practice cluster.
type Lab struct {
	Manifests fs.FS     // lab manifests: *.yaml at the top level, extras under addons/
	Out       io.Writer // progress messages
	Run       Runner
	LookPath  func(tool string) (string, error)
}

// New returns a Lab that runs the real kind and kubectl.
func New(manifests fs.FS, out io.Writer) *Lab {
	return &Lab{Manifests: manifests, Out: out, Run: runCommand, LookPath: exec.LookPath}
}

func runCommand(ctx context.Context, stdin, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (l *Lab) say(format string, args ...any) {
	fmt.Fprintf(l.Out, format+"\n", args...)
}

// run runs a command and turns a failure into an error carrying its output.
func (l *Lab) run(ctx context.Context, stdin, name string, args ...string) (string, error) {
	output, err := l.Run(ctx, stdin, name, args...)
	if err != nil {
		return output, fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, strings.TrimSpace(output))
	}
	return output, nil
}

func (l *Lab) requireTools() error {
	for tool, hint := range map[string]string{
		"kind":    "https://kind.sigs.k8s.io/docs/user/quick-start/#installation",
		"kubectl": "https://kubernetes.io/docs/tasks/tools/",
	} {
		if _, err := l.LookPath(tool); err != nil {
			return fmt.Errorf("%s is not installed or not in PATH; see %s", tool, hint)
		}
	}
	return nil
}

func (l *Lab) clusterExists(ctx context.Context) (bool, error) {
	output, err := l.run(ctx, "", "kind", "get", "clusters")
	if err != nil {
		return false, err
	}
	for _, name := range strings.Fields(output) {
		if name == ClusterName {
			return true, nil
		}
	}
	return false, nil
}

// manifestsIn returns the contents of the YAML files directly inside dir, in
// name order.
func (l *Lab) manifestsIn(dir string) (names []string, contents map[string]string, err error) {
	entries, err := fs.ReadDir(l.Manifests, dir)
	if err != nil {
		return nil, nil, err
	}
	contents = map[string]string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		path := entry.Name()
		if dir != "." {
			path = dir + "/" + entry.Name()
		}
		data, err := fs.ReadFile(l.Manifests, path)
		if err != nil {
			return nil, nil, err
		}
		names = append(names, path)
		contents[path] = string(data)
	}
	sort.Strings(names)
	return names, contents, nil
}

func (l *Lab) apply(ctx context.Context, dir string) error {
	names, contents, err := l.manifestsIn(dir)
	if err != nil {
		return err
	}
	for _, name := range names {
		if _, err := l.run(ctx, contents[name], "kubectl", "--context", Context, "apply", "-f", "-"); err != nil {
			return fmt.Errorf("applying %s: %w", name, err)
		}
	}
	return nil
}

// installGatewayAPI installs the Gateway API resource definitions from the
// project's release page, then the lab gateway class.
func (l *Lab) installGatewayAPI(ctx context.Context) error {
	url := "https://github.com/kubernetes-sigs/gateway-api/releases/download/" + GatewayAPIVersion + "/standard-install.yaml"
	if _, err := l.run(ctx, "", "kubectl", "--context", Context, "apply", "-f", url); err != nil {
		return err
	}
	wait := append([]string{"--context", Context, "wait", "--for=condition=Established", "--timeout=60s"}, gatewayCRDs...)
	if _, err := l.run(ctx, "", "kubectl", wait...); err != nil {
		return err
	}
	return l.apply(ctx, "addons")
}

// Up creates the cluster if it does not exist, points kubectl at it, and
// applies the lab resources. It is safe to run again.
func (l *Lab) Up(ctx context.Context) error {
	if err := l.requireTools(); err != nil {
		return err
	}

	exists, err := l.clusterExists(ctx)
	if err != nil {
		return fmt.Errorf("listing kind clusters (is Docker running?): %w", err)
	}
	if exists {
		l.say("Cluster %s already exists.", ClusterName)
	} else {
		l.say("Creating cluster %s with 3 nodes. This takes about a minute...", ClusterName)
		if _, err := l.run(ctx, kindConfig, "kind", "create", "cluster",
			"--name", ClusterName, "--image", NodeImage, "--config", "-", "--wait", "120s"); err != nil {
			return err
		}
	}

	if _, err := l.run(ctx, "", "kubectl", "config", "use-context", Context); err != nil {
		return err
	}
	l.say("kubectl now uses context %s.", Context)

	l.say("Applying the lab resources...")
	if err := l.apply(ctx, "."); err != nil {
		return err
	}

	l.say("Installing the Gateway API %s (needs internet)...", GatewayAPIVersion)
	if err := l.installGatewayAPI(ctx); err != nil {
		// Only four networking questions need it; the rest of the lab works.
		l.say("Warning: the Gateway API was not installed, so questions 51 to 54 cannot be solved yet.\n%v\nRun \"kubelearn up\" again when you are online.", err)
	}

	l.say("The lab is ready.")
	return nil
}

// Down deletes the cluster.
func (l *Lab) Down(ctx context.Context) error {
	if _, err := l.LookPath("kind"); err != nil {
		return fmt.Errorf("kind is not installed or not in PATH")
	}
	l.say("Deleting cluster %s...", ClusterName)
	if _, err := l.run(ctx, "", "kind", "delete", "cluster", "--name", ClusterName); err != nil {
		return err
	}
	l.say("Cluster deleted.")
	return nil
}

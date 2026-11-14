package k8s

import (
	"os"
	"path/filepath"
	"testing"
)

func writeKubeconfig(t *testing.T, dir, name, server string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	content := `apiVersion: v1
kind: Config
clusters:
- name: test
  cluster:
    server: ` + server + `
contexts:
- name: test
  context:
    cluster: test
    user: test
current-context: test
users:
- name: test
  user: {}
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// client-go resolves the default ~/.kube/config path once, at program start,
// so these tests always select the file through KUBECONFIG; that keeps the
// developer's real configuration out of the results.

func TestLoadKubeConfigHonoursKUBECONFIG(t *testing.T) {
	custom := writeKubeconfig(t, t.TempDir(), "custom-config", "https://custom.example:6443")
	t.Setenv("KUBECONFIG", custom)

	config, err := LoadKubeConfig()
	if err != nil {
		t.Fatal(err)
	}
	if config.Host != "https://custom.example:6443" {
		t.Errorf("Host = %q, want the server from KUBECONFIG", config.Host)
	}
	// An evaluation of every question must fit in one burst.
	if config.Burst < 100 || config.QPS < 50 {
		t.Errorf("client rate limit too low for an evaluation: QPS %v, burst %d", config.QPS, config.Burst)
	}
}

// KUBECONFIG may list several files; the current context comes from the first
// one that sets it, as with kubectl.
func TestLoadKubeConfigMergesKUBECONFIGList(t *testing.T) {
	dir := t.TempDir()
	first := writeKubeconfig(t, dir, "first", "https://first.example:6443")
	second := writeKubeconfig(t, dir, "second", "https://second.example:6443")
	t.Setenv("KUBECONFIG", first+string(os.PathListSeparator)+second)

	config, err := LoadKubeConfig()
	if err != nil {
		t.Fatal(err)
	}
	if config.Host != "https://first.example:6443" {
		t.Errorf("Host = %q, want the server from the first file", config.Host)
	}
}

func TestLoadKubeConfigFailsWithoutConfiguration(t *testing.T) {
	t.Setenv("KUBECONFIG", filepath.Join(t.TempDir(), "does-not-exist"))
	t.Setenv("KUBERNETES_SERVICE_HOST", "")
	t.Setenv("KUBERNETES_SERVICE_PORT", "")

	if config, err := LoadKubeConfig(); err == nil {
		t.Errorf("expected an error without any kubeconfig, got host %q", config.Host)
	}
}

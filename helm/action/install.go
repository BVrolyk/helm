package action

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/deis/helm/helm/manifest"
	"github.com/deis/helm/helm/model"
)

import (
	"github.com/deis/helm/helm/log"
)

// Install loads a chart into Kubernetes.
//
// If the chart is not found in the workspace, it is fetched and then installed.
//
// During install, manifests are sent to Kubernetes in the following order:
//
//	- Namespaces
// 	- Secrets
// 	- Volumes
// 	- Services
// 	- Pods
// 	- ReplicationControllers
func Install(chart, home, namespace string) {
	if !chartInstalled(chart, home) {
		log.Info("No installed chart named %q. Installing now.", chart)
		Fetch(chart, chart, home)
	}

	cd := filepath.Join(home, WorkspaceChartPath, chart)
	c, err := model.Load(cd)
	if err != nil {
		log.Die("Failed to load chart: %s", err)
	}

	if err := uploadManifests(c); err != nil {
		log.Die("Failed to upload manifests: %s", err)
	}
}

// uploadManifests sends manifests to Kubectl in a particular order.
func uploadManifests(c *model.Chart) error {
	// The ordering is significant.
	// TODO: Right now, we force version v1. We could probably make this more
	// flexible if there is a use case.
	for _, o := range c.Namespaces {
		b, err := manifest.MarshalJSON(o, "v1")
		if err != nil {
			return err
		}
		if err := kubectlCreate(b); err != nil {
			return err
		}
	}
	for _, o := range c.Secrets {
		b, err := manifest.MarshalJSON(o, "v1")
		if err != nil {
			return err
		}
		if err := kubectlCreate(b); err != nil {
			return err
		}
	}
	for _, o := range c.PersistentVolumes {
		b, err := manifest.MarshalJSON(o, "v1")
		if err != nil {
			return err
		}
		if err := kubectlCreate(b); err != nil {
			return err
		}
	}
	for _, o := range c.Services {
		b, err := manifest.MarshalJSON(o, "v1")
		if err != nil {
			return err
		}
		if err := kubectlCreate(b); err != nil {
			return err
		}
	}
	for _, o := range c.Pods {
		b, err := manifest.MarshalJSON(o, "v1")
		if err != nil {
			return err
		}
		if err := kubectlCreate(b); err != nil {
			return err
		}
	}
	for _, o := range c.ReplicationControllers {
		b, err := manifest.MarshalJSON(o, "v1")
		if err != nil {
			return err
		}
		if err := kubectlCreate(b); err != nil {
			return err
		}
	}
	return nil
}

// Check by chart directory name whether a chart is installed.
//
// This does NOT check the Chart.yaml file.
func chartInstalled(chart, home string) bool {
	p := filepath.Join(home, WorkspaceChartPath, chart, "Chart.yaml")
	log.Debug("Looking for %q", p)
	if fi, err := os.Stat(p); err != nil || fi.IsDir() {
		log.Debug("No chart: %s", err)
		return false
	}
	return true
}

// kubectlCreate calls `kubectl create` and sends the data via Stdin.
func kubectlCreate(data []byte) error {
	a := []string{"create", "-f", "-"}
	c := exec.Command("kubectl", a...)
	in, err := c.StdinPipe()
	if err != nil {
		return err
	}

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Start(); err != nil {
		return err
	}

	log.Info("File: %s", string(data))
	in.Write(data)
	in.Close()

	return c.Wait()
}

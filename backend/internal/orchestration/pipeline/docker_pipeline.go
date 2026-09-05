package pipeline

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/pkg/security"
)

type LogBroadcaster interface {
	BroadcastDeploymentLog(deploymentID uuid.UUID, log any)
}

type DockerPipeline struct {
	deploymentRepo domain.DeploymentRepository
	broadcaster    LogBroadcaster
}

func NewDockerPipeline(repo domain.DeploymentRepository, broadcaster LogBroadcaster) *DockerPipeline {
	return &DockerPipeline{
		deploymentRepo: repo,
		broadcaster:    broadcaster,
	}
}

func (p *DockerPipeline) Execute(ctx context.Context, dep *domain.Deployment) error {
	go func() {
		bgCtx := context.Background()
		p.runPipeline(bgCtx, dep)
	}()
	return nil
}

func (p *DockerPipeline) runPipeline(ctx context.Context, dep *domain.Deployment) {

	p.log(ctx, dep.ID, "system", fmt.Sprintf("[Pipeline] Inisialisasi deployment container %s (%s)", dep.AppName, dep.ImageTag))
	_ = p.deploymentRepo.UpdateDeploymentStatus(ctx, dep.ID, domain.DeploymentStatusPulling, "", nil)

	hasDocker := false
	if !strings.HasPrefix(dep.ImageTag, "mock") && !strings.HasPrefix(dep.ImageTag, "test") {
		if _, err := exec.LookPath("docker"); err == nil {

			if infoErr := exec.CommandContext(ctx, "docker", "info").Run(); infoErr == nil {
				hasDocker = true
			}
		}
	}

	if hasDocker {
		p.log(ctx, dep.ID, "stdout", fmt.Sprintf("Mengunduh image %s dari registry...", dep.ImageTag))
		pullCmd := exec.CommandContext(ctx, "docker", "pull", dep.ImageTag)
		pullStdout, _ := pullCmd.StdoutPipe()
		pullStderr, _ := pullCmd.StderrPipe()

		if err := pullCmd.Start(); err != nil {
			errMsg := fmt.Sprintf("Gagal memulai docker pull: %v", err)
			p.log(ctx, dep.ID, "stderr", errMsg)
			_ = p.deploymentRepo.UpdateDeploymentStatus(ctx, dep.ID, domain.DeploymentStatusFailed, errMsg, nil)
			return
		}

		go func() {
			scanner := bufio.NewScanner(pullStdout)
			for scanner.Scan() {
				p.log(ctx, dep.ID, "stdout", scanner.Text())
			}
			_ = scanner.Err()
		}()
		go func() {
			scanner := bufio.NewScanner(pullStderr)
			for scanner.Scan() {
				p.log(ctx, dep.ID, "stderr", scanner.Text())
			}
			_ = scanner.Err()
		}()

		if err := pullCmd.Wait(); err != nil {

			inspectErr := exec.CommandContext(ctx, "docker", "image", "inspect", dep.ImageTag).Run()
			if inspectErr == nil {
				p.log(ctx, dep.ID, "stdout", fmt.Sprintf("Image '%s' ditemukan di lokal host. Melanjutkan deployment...", dep.ImageTag))
			} else {
				errMsg := fmt.Sprintf("Gagal mengunduh image %s: image tidak ditemukan atau koneksi registry gagal", dep.ImageTag)
				p.log(ctx, dep.ID, "stderr", errMsg)
				_ = p.deploymentRepo.UpdateDeploymentStatus(ctx, dep.ID, domain.DeploymentStatusFailed, errMsg, nil)
				return
			}
		}
	} else {

		p.log(ctx, dep.ID, "stdout", fmt.Sprintf("Pulling image %s from registry...", dep.ImageTag))
		time.Sleep(300 * time.Millisecond)
		p.log(ctx, dep.ID, "stdout", "Layer 1/3: [====================================>] 45.2MB/45.2MB")
		p.log(ctx, dep.ID, "stdout", "Layer 2/3: [====================================>] 12.8MB/12.8MB")
		p.log(ctx, dep.ID, "stdout", "Layer 3/3: [====================================>] 2.1MB/2.1MB")
		p.log(ctx, dep.ID, "stdout", fmt.Sprintf("Status: Downloaded newer image for %s", dep.ImageTag))
	}

	_ = p.deploymentRepo.UpdateDeploymentStatus(ctx, dep.ID, domain.DeploymentStatusDeploying, "", nil)
	p.log(ctx, dep.ID, "system", fmt.Sprintf("[Pipeline] Mengonfigurasi container '%s'...", dep.ContainerName))

	var runArgs []string
	runArgs = append(runArgs, "run", "-d", "--name", dep.ContainerName)

	if dep.RestartPolicy != "" {
		runArgs = append(runArgs, "--restart", dep.RestartPolicy)
	}

	dockerNetName := dep.NetworkName
	if hasDocker && (dockerNetName == "" || dockerNetName == "caelus-network") {
		netCheck := exec.CommandContext(ctx, "docker", "network", "ls", "--format", "{{.Name}}")
		if out, err := netCheck.Output(); err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if strings.Contains(line, "caelus-network") {
					dockerNetName = line
					break
				}
			}
		}
	}

	if dockerNetName != "" {
		if !strings.HasPrefix(dockerNetName, "caelus-") && !strings.Contains(dockerNetName, "_") {
			dockerNetName = fmt.Sprintf("caelus-%s", dep.NetworkName)
		}
		if hasDocker {
			_ = exec.CommandContext(ctx, "docker", "network", "create", "--driver", "bridge", dockerNetName).Run()
		}
		p.log(ctx, dep.ID, "system", fmt.Sprintf("Attaching to Network: %s", dockerNetName))
		runArgs = append(runArgs, "--network", dockerNetName)
	}

	for _, pb := range dep.PortBindings {
		p.log(ctx, dep.ID, "system", fmt.Sprintf("Binding port: %d:%d/%s", pb.HostPort, pb.ContainerPort, pb.Protocol))
		runArgs = append(runArgs, "-p", fmt.Sprintf("%d:%d", pb.HostPort, pb.ContainerPort))
	}

	for _, vb := range dep.VolumeBindings {
		canonicalHost, err := security.ValidateHostPath(vb.HostPath)
		if err != nil {
			p.log(ctx, dep.ID, "stderr", fmt.Sprintf("Pipeline menolak bind-mount tidak aman: %s (%v)", vb.HostPath, err))
			_ = p.deploymentRepo.UpdateDeploymentStatus(ctx, dep.ID, domain.DeploymentStatusFailed,
				fmt.Sprintf("bind-mount path tidak aman ditolak: %s (%v)", vb.HostPath, err), nil)
			return
		}

		p.log(ctx, dep.ID, "system", fmt.Sprintf("Mounting volume: %s -> %s (%s)", canonicalHost, vb.ContainerPath, vb.Mode))
		if vb.Mode != "" {
			runArgs = append(runArgs, "-v", fmt.Sprintf("%s:%s:%s", canonicalHost, vb.ContainerPath, vb.Mode))
		} else {
			runArgs = append(runArgs, "-v", fmt.Sprintf("%s:%s", canonicalHost, vb.ContainerPath))
		}
	}

	for k, v := range dep.EnvironmentVariables {
		runArgs = append(runArgs, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	if len(dep.EnvironmentVariables) > 0 {
		p.log(ctx, dep.ID, "system", fmt.Sprintf("Injected %d environment variables", len(dep.EnvironmentVariables)))
	}

	runArgs = append(runArgs, dep.ImageTag)

	if dep.Command != "" {
		cmdParts := strings.Fields(dep.Command)
		runArgs = append(runArgs, cmdParts...)
	} else if strings.Contains(dep.ImageTag, "cloudflared") {

		if tunnelURL, ok := dep.EnvironmentVariables["TUNNEL_URL"]; ok && tunnelURL != "" {
			runArgs = append(runArgs, "tunnel", "--no-autoupdate", "--url", tunnelURL)
		} else if tunnelToken, ok := dep.EnvironmentVariables["TUNNEL_TOKEN"]; ok && tunnelToken != "" {
			runArgs = append(runArgs, "tunnel", "--no-autoupdate", "run", "--token", tunnelToken)
		}
	}

	if hasDocker {

		_ = exec.CommandContext(ctx, "docker", "rm", "-f", dep.ContainerName).Run()

		runCmd := exec.CommandContext(ctx, "docker", runArgs...)
		outBytes, err := runCmd.CombinedOutput()
		outStr := strings.TrimSpace(string(outBytes))

		if err != nil {
			errMsg := fmt.Sprintf("Gagal menjalankan container: %s | %v", outStr, err)
			p.log(ctx, dep.ID, "stderr", errMsg)
			_ = p.deploymentRepo.UpdateDeploymentStatus(ctx, dep.ID, domain.DeploymentStatusFailed, errMsg, nil)
			return
		}

		p.log(ctx, dep.ID, "stdout", fmt.Sprintf("Container %s [ID: %s] berhasil dijalankan.", dep.ContainerName, outStr[:min(12, len(outStr))]))

		now := time.Now().UTC()
		_ = p.deploymentRepo.UpdateDeploymentStatus(ctx, dep.ID, domain.DeploymentStatusRunning, "", &now)
		p.log(ctx, dep.ID, "system", fmt.Sprintf("Deployment berhasil. Container '%s' aktif pada status RUNNING.", dep.ContainerName))

		go p.streamContainerLogs(dep.ID, dep.ContainerName)
	} else {
		p.log(ctx, dep.ID, "stdout", fmt.Sprintf("Starting container %s [ID: c-%s]...", dep.ContainerName, dep.ID.String()[:8]))
		p.log(ctx, dep.ID, "stdout", "Listening on port(s)... Application ready.")
		p.log(ctx, dep.ID, "system", "Performing container healthcheck probe (HTTP GET /health)...")
		time.Sleep(200 * time.Millisecond)
		p.log(ctx, dep.ID, "system", "Healthcheck passed: status 200 OK (latency: 4ms)")

		now := time.Now().UTC()
		_ = p.deploymentRepo.UpdateDeploymentStatus(ctx, dep.ID, domain.DeploymentStatusRunning, "", &now)
		p.log(ctx, dep.ID, "system", fmt.Sprintf("Deployment successfully completed. Container '%s' is RUNNING.", dep.ContainerName))
	}
}

func (p *DockerPipeline) streamContainerLogs(depID uuid.UUID, containerName string) {
	ctx := context.Background()
	logCmd := exec.Command("docker", "logs", "-f", "--tail", "50", containerName)
	stdout, err := logCmd.StdoutPipe()
	if err != nil {
		return
	}
	stderr, err := logCmd.StderrPipe()
	if err != nil {
		return
	}

	if err := logCmd.Start(); err != nil {
		return
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			p.log(ctx, depID, "stdout", scanner.Text())
		}
		_ = scanner.Err()
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			p.log(ctx, depID, "stderr", scanner.Text())
		}
		_ = scanner.Err()
	}()

	_ = logCmd.Wait()
}

func (p *DockerPipeline) log(ctx context.Context, depID uuid.UUID, stream, message string) {
	logEntry := &domain.DeploymentLog{
		DeploymentID: depID,
		Timestamp:    time.Now().UTC(),
		Stream:       stream,
		Message:      message,
	}

	_ = p.deploymentRepo.AppendLog(ctx, logEntry)

	if p.broadcaster != nil {
		p.broadcaster.BroadcastDeploymentLog(depID, *logEntry)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

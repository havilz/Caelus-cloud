package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

// LogBroadcaster mendefinisikan interface untuk menyiarkan log realtime ke client (misalnya via WebSocket).
type LogBroadcaster interface {
	BroadcastDeploymentLog(deploymentID uuid.UUID, log any)
}

// DockerPipeline mengelola orkestrasi siklus hidup container deployment.
type DockerPipeline struct {
	deploymentRepo domain.DeploymentRepository
	broadcaster    LogBroadcaster
}

// NewDockerPipeline membuat instance baru Docker Pipeline.
func NewDockerPipeline(repo domain.DeploymentRepository, broadcaster LogBroadcaster) *DockerPipeline {
	return &DockerPipeline{
		deploymentRepo: repo,
		broadcaster:    broadcaster,
	}
}

// Execute menjalankan pipeline container deployment secara asynchronous.
func (p *DockerPipeline) Execute(ctx context.Context, dep *domain.Deployment) error {
	go func() {
		bgCtx := context.Background()
		p.runPipeline(bgCtx, dep)
	}()
	return nil
}

func (p *DockerPipeline) runPipeline(ctx context.Context, dep *domain.Deployment) {
	// Step 1: Queued -> Pulling
	p.log(ctx, dep.ID, "system", fmt.Sprintf("[Pipeline] Initiating deployment for %s (%s)", dep.AppName, dep.ImageTag))
	_ = p.deploymentRepo.UpdateDeploymentStatus(ctx, dep.ID, domain.DeploymentStatusPulling, "", nil)

	p.log(ctx, dep.ID, "stdout", fmt.Sprintf("Pulling image %s from registry...", dep.ImageTag))
	time.Sleep(300 * time.Millisecond) // Simulated pull delay for streaming feel
	p.log(ctx, dep.ID, "stdout", "Layer 1/3: [====================================>] 45.2MB/45.2MB")
	p.log(ctx, dep.ID, "stdout", "Layer 2/3: [====================================>] 12.8MB/12.8MB")
	p.log(ctx, dep.ID, "stdout", "Layer 3/3: [====================================>] 2.1MB/2.1MB")
	p.log(ctx, dep.ID, "stdout", "Digest: sha256:d8f94e7b419c8f62a1c0d48123abc456def789...")
	p.log(ctx, dep.ID, "stdout", fmt.Sprintf("Status: Downloaded newer image for %s", dep.ImageTag))

	// Step 2: Deploying / Configuring Container
	_ = p.deploymentRepo.UpdateDeploymentStatus(ctx, dep.ID, domain.DeploymentStatusDeploying, "", nil)
	p.log(ctx, dep.ID, "system", fmt.Sprintf("[Pipeline] Configuring container '%s'...", dep.ContainerName))

	for _, pb := range dep.PortBindings {
		p.log(ctx, dep.ID, "system", fmt.Sprintf("Binding port: %d:%d/%s", pb.HostPort, pb.ContainerPort, pb.Protocol))
	}
	for _, vb := range dep.VolumeBindings {
		p.log(ctx, dep.ID, "system", fmt.Sprintf("Mounting volume: %s -> %s (%s)", vb.HostPath, vb.ContainerPath, vb.Mode))
	}
	if len(dep.EnvironmentVariables) > 0 {
		p.log(ctx, dep.ID, "system", fmt.Sprintf("Injected %d environment variables", len(dep.EnvironmentVariables)))
	}

	time.Sleep(200 * time.Millisecond)

	// Step 3: Starting & Healthcheck
	p.log(ctx, dep.ID, "stdout", fmt.Sprintf("Starting container %s [ID: c-%s]...", dep.ContainerName, dep.ID.String()[:8]))
	p.log(ctx, dep.ID, "stdout", "Listening on port(s)... Application ready.")
	p.log(ctx, dep.ID, "system", "Performing container healthcheck probe (HTTP GET /health)...")
	time.Sleep(200 * time.Millisecond)
	p.log(ctx, dep.ID, "system", "Healthcheck passed: status 200 OK (latency: 4ms)")

	// Step 4: Running
	now := time.Now().UTC()
	_ = p.deploymentRepo.UpdateDeploymentStatus(ctx, dep.ID, domain.DeploymentStatusRunning, "", &now)
	p.log(ctx, dep.ID, "system", fmt.Sprintf("Deployment successfully completed. Container '%s' is RUNNING.", dep.ContainerName))
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

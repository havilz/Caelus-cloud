package helper

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/havilz/caelus-cloud/backend/internal/delivery/http/response"
)

func HandleAgentBinDownload(w http.ResponseWriter, r *http.Request) {
	candidatePaths := []string{
		filepath.Join("agent", "bin", "caelus-agent"),
		filepath.Join("..", "agent", "bin", "caelus-agent"),
		filepath.Join("..", "..", "agent", "bin", "caelus-agent"),
		filepath.Join("/opt", "caelus", "caelus-agent"),
	}

	var targetBin string
	for _, path := range candidatePaths {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			targetBin = path
			break
		}
	}

	if targetBin == "" {
		response.Error(w, http.StatusNotFound, "caelus-agent binary not compiled on API server (Run 'make build-agent')", nil)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=caelus-agent")
	http.ServeFile(w, r, targetBin)
}

// HandleAgentInstallScript serves the shell installation script for Caelus Agent.
func HandleAgentInstallScript(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(GetAgentInstallScript()))
}

// GetAgentInstallScript returns the bash script for installing Caelus Agent on target VPS/servers.
func GetAgentInstallScript() string {
	return `#!/usr/bin/env bash
set -e

SERVER_ID=""
AGENT_SECRET=""
API_ENDPOINT="http://localhost:8080"

while [[ "$#" -gt 0 ]]; do
    case $1 in
        --server-id=*) SERVER_ID="${1#*=}" ;;
        --server-id) SERVER_ID="$2"; shift ;;
        --secret=*) AGENT_SECRET="${1#*=}" ;;
        --secret) AGENT_SECRET="$2"; shift ;;
        --api=*|--endpoint=*) API_ENDPOINT="${1#*=}" ;;
        --api|--endpoint) API_ENDPOINT="$2"; shift ;;
        *) echo "Unknown parameter: $1"; exit 1 ;;
    esac
    shift
done

if [ -z "$SERVER_ID" ] || [ -z "$AGENT_SECRET" ]; then
    echo "Usage: curl -sSL https://caelus.cloud/install.sh | bash -s -- --server-id <ID> --secret <SECRET> [--api <URL>]"
    exit 1
fi

INSTALL_DIR="/opt/caelus"
mkdir -p "$INSTALL_DIR"

echo "-> Stopping previous service if running..."
systemctl stop caelus-agent 2>/dev/null || true

echo "-> Downloading latest agent binary..."
curl -sSL "$API_ENDPOINT/agent-bin" -o "$INSTALL_DIR/caelus-agent.tmp"
mv -f "$INSTALL_DIR/caelus-agent.tmp" "$INSTALL_DIR/caelus-agent"
chmod +x "$INSTALL_DIR/caelus-agent"

echo "-> Generating agent configuration..."
cat <<EOF > "$INSTALL_DIR/agent.env"
SERVER_ID=$SERVER_ID
AGENT_SECRET=$AGENT_SECRET
API_ENDPOINT=$API_ENDPOINT
COLLECTION_INTERVAL_SEC=5
CAELUS_SERVER_ID=$SERVER_ID
CAELUS_AGENT_SECRET=$AGENT_SECRET
CAELUS_API_ENDPOINT=$API_ENDPOINT
CAELUS_INTERVAL=5s
EOF

if [ -n "$ALL_PROXY" ]; then
    echo "ALL_PROXY=$ALL_PROXY" >> "$INSTALL_DIR/agent.env"
    echo "HTTP_PROXY=$ALL_PROXY" >> "$INSTALL_DIR/agent.env"
    echo "HTTPS_PROXY=$ALL_PROXY" >> "$INSTALL_DIR/agent.env"
fi

echo "-> Registering systemd service..."
cat <<EOF > /etc/systemd/system/caelus-agent.service
[Unit]
Description=Caelus Cloud Monitoring & Telemetry Agent
After=network.target

[Service]
Type=simple
EnvironmentFile=$INSTALL_DIR/agent.env
ExecStart=$INSTALL_DIR/caelus-agent
Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload 2>/dev/null || true
systemctl enable caelus-agent 2>/dev/null || true
systemctl restart caelus-agent 2>/dev/null || true

echo "=== Caelus Cloud Agent Installed and Running Successfully! ==="
`
}

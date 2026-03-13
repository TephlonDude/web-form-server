#!/usr/bin/env bash
# deploy.sh — Build locally, transfer image, and start on remote host.
# Usage: ./deploy.sh [--rebuild]
#   --rebuild  force Docker to ignore build cache

set -euo pipefail

# ── Config ────────────────────────────────────────────────────────────────────
REMOTE_HOST="192.168.202.206"
REMOTE_USER="jason"
REMOTE_DIR="/home/jason/web-form-server"
IMAGE_NAME="web-form-server"
IMAGE_TAG="latest"
TARBALL="/tmp/web-form-server.tar"
SSH_KEY="$HOME/.ssh/t5n_api_server"
SSH_OPTS="-i ${SSH_KEY} -o StrictHostKeyChecking=no"

BUILD_FLAGS=""
if [[ "${1:-}" == "--rebuild" ]]; then
  BUILD_FLAGS="--no-cache"
  echo "==> Cache disabled (--rebuild)"
fi

# ── 1. Build ──────────────────────────────────────────────────────────────────
echo "==> Building Docker image..."
docker build $BUILD_FLAGS -t "${IMAGE_NAME}:${IMAGE_TAG}" .

# ── 2. Save ───────────────────────────────────────────────────────────────────
echo "==> Saving image to ${TARBALL}..."
docker save "${IMAGE_NAME}:${IMAGE_TAG}" -o "${TARBALL}"

IMAGE_SIZE=$(du -sh "${TARBALL}" | cut -f1)
echo "    Image size: ${IMAGE_SIZE}"

# ── 3. Prepare remote directory structure ────────────────────────────────────
echo "==> Preparing remote directories..."
ssh $SSH_OPTS "${REMOTE_USER}@${REMOTE_HOST}" \
  "mkdir -p ${REMOTE_DIR}/web-config ${REMOTE_DIR}/data/submissions"

# ── 4. Transfer files ────────────────────────────────────────────────────────
echo "==> Transferring image (this may take a moment)..."
scp $SSH_OPTS "${TARBALL}" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/web-form-server.tar"

echo "==> Transferring docker-compose.yml..."
scp $SSH_OPTS docker-compose.yml "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/"

echo "==> Transferring web-config/..."
scp $SSH_OPTS -r web-config/ "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/"

# ── 5. Load image and restart container ──────────────────────────────────────
echo "==> Loading image on remote and restarting container..."
ssh $SSH_OPTS "${REMOTE_USER}@${REMOTE_HOST}" bash <<EOF
  set -e
  cd "${REMOTE_DIR}"

  # Load the new image
  docker load -i web-form-server.tar
  rm web-form-server.tar

  # Bring up (creates or recreates container as needed)
  docker compose down --remove-orphans 2>/dev/null || true
  docker compose up -d

  # Wait a moment and check health
  sleep 3
  docker compose ps
EOF

# ── 6. Clean up local tarball ────────────────────────────────────────────────
echo "==> Cleaning up local tarball..."
rm -f "${TARBALL}"

# ── 7. Done ───────────────────────────────────────────────────────────────────
echo ""
echo "================================================"
echo "  Deployment complete!"
echo "  Form:   http://${REMOTE_HOST}:8237"
echo "  Config: http://${REMOTE_HOST}:8237/config"
echo "  Health: http://${REMOTE_HOST}:8237/health"
echo "================================================"

#!/usr/bin/env bash
# deploy.sh — Deploy web-form-server to the remote Docker host.
#
# Usage: ./deploy.sh [--rebuild]
#   --rebuild  force Docker to ignore build cache
#
# Strategy:
#   If local Docker is available: build locally → save → scp → load on remote.
#   Otherwise: copy source files to remote and build there.

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

# ── Detect local Docker ───────────────────────────────────────────────────────
if docker info &>/dev/null; then
  LOCAL_DOCKER=true
  echo "==> Local Docker detected — building locally"
else
  LOCAL_DOCKER=false
  echo "==> Local Docker unavailable — building on remote server"
fi

# ── Prepare remote directories ────────────────────────────────────────────────
echo "==> Preparing remote directories..."
ssh $SSH_OPTS "${REMOTE_USER}@${REMOTE_HOST}" \
  "mkdir -p ${REMOTE_DIR}/templates ${REMOTE_DIR}/web-config ${REMOTE_DIR}/data/submissions"

if $LOCAL_DOCKER; then
  # ── Local build path ────────────────────────────────────────────────────────

  echo "==> Building Docker image..."
  docker build $BUILD_FLAGS -t "${IMAGE_NAME}:${IMAGE_TAG}" .

  echo "==> Saving image to ${TARBALL}..."
  docker save "${IMAGE_NAME}:${IMAGE_TAG}" -o "${TARBALL}"
  echo "    Image size: $(du -sh "${TARBALL}" | cut -f1)"

  echo "==> Transferring image (this may take a moment)..."
  scp $SSH_OPTS "${TARBALL}" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/web-form-server.tar"

  echo "==> Transferring docker-compose.yml and web-config..."
  scp $SSH_OPTS docker-compose.yml "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/"
  scp $SSH_OPTS -r web-config/ "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/"

  echo "==> Loading image on remote and restarting container..."
  ssh $SSH_OPTS "${REMOTE_USER}@${REMOTE_HOST}" bash <<EOF
    set -e
    cd "${REMOTE_DIR}"
    docker load -i web-form-server.tar
    rm web-form-server.tar
    docker compose down --remove-orphans 2>/dev/null || true
    docker compose up -d
    sleep 3
    docker compose ps
EOF

  echo "==> Cleaning up local tarball..."
  rm -f "${TARBALL}"

else
  # ── Remote build path ───────────────────────────────────────────────────────

  echo "==> Copying source files to remote..."
  scp $SSH_OPTS \
    main.go form_loader.go renderer.go submission.go config_handler.go \
    go.mod Dockerfile docker-compose.yml \
    "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/"

  scp $SSH_OPTS \
    templates/form.html templates/submitted.html templates/config.html \
    "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/templates/"

  scp $SSH_OPTS \
    web-config/example.form.toml web-config/example.form.css \
    "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/web-config/"

  echo "==> Building and starting container on remote..."
  ssh $SSH_OPTS "${REMOTE_USER}@${REMOTE_HOST}" bash <<EOF
    set -e
    cd "${REMOTE_DIR}"
    docker compose down --remove-orphans 2>/dev/null || true
    docker compose up -d --build ${BUILD_FLAGS}
    sleep 3
    docker compose ps
EOF

fi

# ── Done ──────────────────────────────────────────────────────────────────────
echo ""
echo "================================================"
echo "  Deployment complete!"
echo "  Form:   http://${REMOTE_HOST}:8237"
echo "  Config: http://${REMOTE_HOST}:8237/config"
echo "  Health: http://${REMOTE_HOST}:8237/health"
echo "================================================"

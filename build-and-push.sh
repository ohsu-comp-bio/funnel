#!/usr/bin/env bash
set -euo pipefail

DOCKERHUB="docker.io/cmgantwerpen/funnel"
ECR="391209680344.dkr.ecr.eu-west-1.amazonaws.com/funnel"
COMMIT=$(git rev-parse --short HEAD)
TAG="multiarch-rev${COMMIT}-develop"

echo "=== Building Funnel image ==="
echo "    Commit    : ${COMMIT}"
echo "    Tag       : ${TAG}"
echo "    Docker Hub: ${DOCKERHUB}:${TAG}"
echo "    ECR       : ${ECR}:${TAG}"
echo ""

# ── Registry login ─────────────────────────────────────────────────────────────
echo "=== Logging in to Docker Hub ==="
# Reads DOCKERHUB_USER / DOCKERHUB_PASS from env, or falls back to interactive.
if [[ -n "${DOCKERHUB_USER:-}" && -n "${DOCKERHUB_PASS:-}" ]]; then
  echo "$DOCKERHUB_PASS" | docker login docker.io -u "$DOCKERHUB_USER" --password-stdin
else
  docker login docker.io
fi

echo "=== Logging in to ECR ==="
aws ecr get-login-password --region eu-west-1 | \
  docker login --username AWS --password-stdin "${ECR%%/*}"

# ── Build & push multiarch image ──────────────────────────────────────────────
echo ""
echo "=== Building multiarch image ==="
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --no-cache \
  -t "${DOCKERHUB}:${TAG}" \
  -t "${ECR}:${TAG}" \
  --push \
  .

echo ""
echo "✅ Done."
echo "   Docker Hub: ${DOCKERHUB}:${TAG}"
echo "   ECR       : ${ECR}:${TAG}"
echo ""
echo "Update funnel-deployment.yaml with tag: ${TAG}"
echo "OVH apply:"
echo "  KUBECONFIG=~/.kube/ovh-tes.yaml kubectl set image deployment/funnel funnel=${DOCKERHUB}:${TAG} -n funnel"
echo "  KUBECONFIG=~/.kube/ovh-tes.yaml kubectl rollout status deployment/funnel -n funnel --timeout=120s"
echo ""
echo "AWS apply:"
echo "  kubectl set image deployment/funnel funnel=${ECR}:${TAG} -n funnel"
echo "  kubectl rollout status deployment/funnel -n funnel --timeout=120s"

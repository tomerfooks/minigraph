#!/usr/bin/env bash
set -euo pipefail

# Build (arm64, ON THE NODE) and deploy the MiniGraph site to the shared k8s
# cluster. The image is built 100% on the server by the shared in-cluster
# builder (devops/build/build.sh -> buildkitd pod): clone from GitHub ->
# ops/Dockerfile (Astro build -> nginx) -> import into containerd. GitHub is
# only the git-clone source, so push before deploying.
#
#   ops/deploy.sh              # build on the node, then deploy
#   ops/deploy.sh --no-build   # deploy manifests / an already-built image only

IMAGE="tomerfooks/minigraph-site:latest"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SSH_USER="${SSH_USER:-ubuntu}"
SSH_HOST="${SSH_HOST:-141.148.71.59}"
SSH_KEY="${SSH_KEY:-$HOME/.ssh/oracle-main.key}"
KUBECTL=("$ROOT/ops/ssh.sh" kubectl)

echo "==> Deploying minigraph-site (minigraph.cite.co.il)"

if [[ "${1:-}" != "--no-build" ]]; then
  BUILD_SH="${BUILD_SH:-$HOME/Projects/devops/build/build.sh}"
  [[ -x "$BUILD_SH" ]] || { echo "ERROR: shared builder not found: $BUILD_SH" >&2; exit 1; }
  GIT_URL="${GIT_URL:-$(git -C "$ROOT" remote get-url origin)}"
  GIT_REF="${GIT_REF:-$(git -C "$ROOT" rev-parse --abbrev-ref HEAD)}"

  # GitOps: the node builds origin/<branch>, so refuse to ship code that isn't there.
  [[ -z "$(git -C "$ROOT" log --oneline @{u}.. 2>/dev/null)" ]] \
    || { echo "ERROR: unpushed commits — push before deploying (the node builds from GitHub)" >&2; exit 1; }
  [[ -z "$(git -C "$ROOT" status --porcelain)" ]] \
    || echo "WARNING: uncommitted changes are NOT on GitHub and won't be in the image" >&2

  echo "==> Building $IMAGE on the node from $GIT_URL#$GIT_REF ..."
  SSH_USER="$SSH_USER" SSH_HOST="$SSH_HOST" SSH_KEY="$SSH_KEY" \
    "$BUILD_SH" "$GIT_URL" "$GIT_REF" ops/Dockerfile "$IMAGE"
else
  echo "==> Skipping build (--no-build)"
fi

echo "==> Applying manifests (ops/k8s.yaml)..."
"${KUBECTL[@]}" apply -f - < "$ROOT/ops/k8s.yaml"

echo "==> Rolling out..."
"${KUBECTL[@]}" rollout restart deployment/minigraph-site -n minigraph
"${KUBECTL[@]}" rollout status deployment/minigraph-site -n minigraph --timeout=180s

echo "==> Done: https://minigraph.cite.co.il"

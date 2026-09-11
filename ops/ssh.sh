#!/usr/bin/env bash
set -euo pipefail

# Run a command on the deploy host (default: the cluster's kubectl box).
# Bundled key is resolved beside the real script, however ssh.sh is invoked.
SRC="${BASH_SOURCE[0]}"
while [[ -L "$SRC" ]]; do
  DIR="$(cd -P "$(dirname "$SRC")" && pwd)"
  SRC="$(readlink "$SRC")"
  [[ "$SRC" != /* ]] && SRC="$DIR/$SRC"
done
SCRIPT_DIR="$(cd -P "$(dirname "$SRC")" && pwd)"
SSH_USER="${SSH_USER:-ubuntu}"
SSH_HOST="${SSH_HOST:-141.148.71.59}"
SSH_KEY="${SSH_KEY:-$HOME/.ssh/oracle-main.key}"
SSH_OPTS="${SSH_OPTS:-}"

[[ -f "$SSH_KEY" ]] || { echo "ssh.sh: key not found: $SSH_KEY" >&2; exit 1; }

# SSH refuses a group/other-readable private key; tighten instead of failing.
if [[ "$(stat -f '%Lp' "$SSH_KEY" 2>/dev/null || stat -c '%a' "$SSH_KEY")" != "600" ]]; then
  chmod 600 "$SSH_KEY"
fi

# shellcheck disable=SC2086
exec ssh -i "$SSH_KEY" \
  -o StrictHostKeyChecking=accept-new \
  -o ConnectTimeout=10 \
  $SSH_OPTS \
  "$SSH_USER@$SSH_HOST" "$@"

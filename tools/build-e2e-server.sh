#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$repo_root/frontend"
pnpm svelte-kit sync
pnpm build

rm -rf "$repo_root/cli/internal/http_server/.client"
cp -r "$repo_root/frontend/build" "$repo_root/cli/internal/http_server/.client"
touch "$repo_root/cli/internal/http_server/.client/.gitkeep"

cd "$repo_root/cli"
CGO_ENABLED="${CGO_ENABLED:-0}" go build \
  -buildvcs=false \
  -tags 'bootstrap test_endpoints' \
  -ldflags '-s -w' \
  -o "$repo_root/frontend/e2e/fixtures/bin/chatto" \
  .

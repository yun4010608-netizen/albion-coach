#!/bin/sh
# Restart app replicas one at a time; the next starts only after the previous is healthy again.
set -eu
cd "$(dirname "$0")/.."
for c in $(docker compose ps -q app); do
  name=$(docker inspect -f '{{.Name}}' "$c")
  echo "restarting $name"
  docker restart -t 210 "$c" >/dev/null
  until [ "$(docker inspect -f '{{.State.Health.Status}}' "$c")" = healthy ]; do sleep 1; done
  echo "$name healthy"
done

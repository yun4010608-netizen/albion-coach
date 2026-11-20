#!/bin/sh
# Zero-downtime image update: start a new replica set next to the old one, wait until
# every new replica is healthy, then stop the old ones (each drains in-flight requests).
set -eu
cd "$(dirname "$0")/.."
want=$(docker compose config --format json | python3 -c 'import json,sys; print(json.load(sys.stdin)["services"]["app"]["deploy"]["replicas"])')
old=$(docker compose ps -q app | tr '\n' ' ')
docker compose build app
docker compose up -d --no-deps --no-recreate --scale app=$((want * 2)) app
for c in $(docker compose ps -q app); do
  case " $old " in *" $c "*) continue ;; esac
  until [ "$(docker inspect -f '{{.State.Health.Status}}' "$c")" = healthy ]; do sleep 1; done
  echo "new replica $(docker inspect -f '{{.Name}}' "$c") healthy"
done
for c in $old; do
  echo "draining $(docker inspect -f '{{.Name}}' "$c")"
  docker stop -t 210 "$c" >/dev/null && docker rm "$c" >/dev/null
done
docker compose ps app

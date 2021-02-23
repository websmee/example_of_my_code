#!/bin/sh
# wait until consul is up
# then inject all key/value files as found in the init directory
set -ue

let "timeout = $(date +%s) + 15"

echo "kv_loader waiting for consul"
while ! curl -f -s http://localhost:8500/v1/status/leader | grep "[0-9]:[0-9]"; do
  if [ $(date +%s) -gt $timeout ]; then echo "kv_loader timeout"; exit 1; fi
  sleep 1
  echo "kv_loader still waiting"
done

echo "kv_loader get/put from $INIT_CONSUL_KV_DIR"
cd $INIT_CONSUL_KV_DIR
for json_file in $(ls *.json); do
  key=$(echo $json_file | sed -e 's/.json$//')
  echo "kv_loader loading $key from $json_file"
  consul kv get $key >/dev/null && echo "kv_loader $key already loaded" || consul kv put $key @$json_file
done
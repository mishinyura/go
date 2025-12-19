#!/usr/bin/env bash

set -e

cd "$(dirname "$0")/.."

protoc \
  -I api/proto \
  --go_out internal/pb --go_opt paths=source_relative \
  --go-grpc_out internal/pb --go-grpc_opt paths=source_relative \
  $(find api/proto -name '*.proto')

echo "Protobuf generated into internal/pb"

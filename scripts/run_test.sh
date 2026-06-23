#!/bin/bash
set -e

echo "Running all tests with coverage..."

go test ./... \
  -v \
  -cover \
  -coverprofile=coverage.out \
  -race \
  -timeout 30s

echo ""
echo "Coverage summary:"
go tool cover -func=coverage.out | grep total

echo ""
echo "Done."

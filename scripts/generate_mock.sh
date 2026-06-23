#!/bin/bash
set -e

echo "Generating mocks..."

mockgen \
  -source=internal/domain/auth/repository/user_repository.go \
  -destination=internal/mock/mock_user_repository.go \
  -package=mock

mockgen \
  -source=internal/domain/auth/usecase/auth_usecase.go \
  -destination=internal/mock/mock_auth_usecase.go \
  -package=mock

mockgen \
  -source=internal/domain/todo/repository/todo_repository.go \
  -destination=internal/mock/mock_todo_repository.go \
  -package=mock

mockgen \
  -source=internal/domain/todo/usecase/todo_usecase.go \
  -destination=internal/mock/mock_todo_usecase.go \
  -package=mock

mockgen \
  -source=pkg/jwt/jwt.go \
  -destination=internal/mock/mock_jwt_service.go \
  -package=mock

echo "All mocks generated successfully."

.PHONY: docs
docs:
	@echo "🔄 Generating Swagger docs..."
	@go install github.com/swaggo/swag/cmd/swag@v1.16.4
	@swag init -g cmd/server/main.go -o docs --parseDependency
	@echo "✅ Docs updated"

.PHONY: lint
lint:
	@echo "🔍 Running linter..."
	@golangci-lint run --timeout=5m

.PHONY: test
test:
	@echo "🧪 Running tests..."
	@go test ./... -race -cover

.PHONY: ci
ci: lint test docs  # Полный прогон как в CI
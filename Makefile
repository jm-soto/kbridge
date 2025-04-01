
help: ## Show this help
	@echo "Usage: make [command]"
	@echo "Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the project
	go build -o kbridge main.go

# Run the project
run: ## Run the project
	go run main.go

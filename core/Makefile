# Variáveis
BINARY_NAME=leonida-core
PLUGIN_DIR=plugins
EXAMPLES_DIR=examples

# Cores para output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m # No Color

build: ## Compila a aplicação
	@echo "$(GREEN)Compilando Leonida Core...$(NC)"
	@go build -o $(BINARY_NAME) main.go
	@echo "$(GREEN)Compilação concluída: $(BINARY_NAME)$(NC)"
	@echo "$(GREEN)Movendo para build/$(BINARY_NAME)...$(NC)"
	@mkdir -p build
	@mv $(BINARY_NAME) build/$(BINARY_NAME)

run: ## Executa a aplicação
	@echo "$(GREEN)Executando Leonida Core...$(NC)"
	@go run main.go

test: ## Executa os testes
	@echo "$(GREEN)Executando testes...$(NC)"
	@go test -v ./...

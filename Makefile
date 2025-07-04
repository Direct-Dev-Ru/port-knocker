.PHONY: build build-linux build-windows clean test help

# Переменные
BINARY_NAME=port-knocker
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -s -w"

# Цвета для вывода
GREEN=\033[0;32m
NC=\033[0m # No Color

help: ## Показать справку
	@echo "$(GREEN)Port Knocker - Утилита для port knocking$(NC)"
	@echo ""
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}'

build: ## Собрать для текущей платформы
	@echo "$(GREEN)Сборка для текущей платформы...$(NC)"
	go build ${LDFLAGS} -o ${BINARY_NAME} .

build-linux: ## Собрать для Linux (amd64)
	@echo "$(GREEN)Сборка для Linux (amd64)...$(NC)"
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-linux-amd64 .

build-windows: ## Собрать для Windows (amd64)
	@echo "$(GREEN)Сборка для Windows (amd64)...$(NC)"
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-windows-amd64.exe .

build-all: build-linux build-windows ## Собрать для всех платформ

test: ## Запустить тесты
	@echo "$(GREEN)Запуск тестов...$(NC)"
	go test -v ./...

clean: ## Очистить собранные файлы
	@echo "$(GREEN)Очистка...$(NC)"
	rm -f ${BINARY_NAME}*
	rm -f *.exe

install: build ## Установить в систему
	@echo "$(GREEN)Установка...$(NC)"
	sudo cp ${BINARY_NAME} /usr/local/bin/

uninstall: ## Удалить из системы
	@echo "$(GREEN)Удаление...$(NC)"
	sudo rm -f /usr/local/bin/${BINARY_NAME}

deps: ## Установить зависимости
	@echo "$(GREEN)Установка зависимостей...$(NC)"
	go mod tidy
	go mod download

example-encrypt: ## Пример шифрования конфигурации
	@echo "$(GREEN)Шифрование примера конфигурации...$(NC)"
	./${BINARY_NAME} encrypt -c examples/config.yaml -o examples/config.encrypted -k examples/key.txt

example-decrypt: ## Пример расшифровки зашифрованной конфигурации
	@echo "$(GREEN)Расшифровка зашифрованной конфигурации...$(NC)"
	./${BINARY_NAME} decrypt -c examples/config.encrypted -o examples/config.decrypted.yaml -k examples/key.txt

example-encrypt-alt: ## Пример шифрования с опцией -i
	@echo "$(GREEN)Шифрование с опцией -i...$(NC)"
	./${BINARY_NAME} encrypt -i examples/config.yaml -o examples/config.encrypted -k examples/key.txt

example-run: ## Пример запуска с обычной конфигурацией
	@echo "$(GREEN)Запуск с обычной конфигурацией...$(NC)"
	./${BINARY_NAME} -c examples/config.yaml -v

release-tag: ## Создать git tag для release (например: make release-tag VERSION=v1.0.0)
	@if [ -z "$(VERSION)" ]; then \
		echo "Использование: make release-tag VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "$(GREEN)Создание тега $(VERSION)...$(NC)"
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

check-git: ## Проверить git статус перед коммитом
	@echo "$(GREEN)Проверка git статуса...$(NC)"
	git status
	@echo ""
	@echo "$(GREEN)Файлы в .gitignore:$(NC)"
	@echo "- *knock*.yaml (конфигурации с 'knock' в имени)"
	@echo "- *.encrypted (зашифрованные файлы)"
	@echo "- Бинарные файлы и ключи" 
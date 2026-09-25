include .env
export

export PROJECT_ROOT=$(shell pwd)

.DEFAULT_GOAL=help

env-up: ## env: Запустить окружение проекта
	@docker compose up -d todo-postgres todo-redis todo-rabbitmq

env-down: ## env: Остановить окружение проекта
	@docker compose down todo-postgres todo-redis todo-rabbitmq

env-cleanup: ## env: Очистить окружение проекта
	@read -p "Очистить все volume файлы окружения? Опасность утери данных. [y/N]: " ans; \
	if [ "$$ans" == "y" ]; then \
		docker compose down -v todo-postgres todo-redis todo-rabbitmq todo-port-forwarder && \
		echo "Файлы окружения очищены"; \
	else \
		echo "Очистка окружения отменена"; \
	fi;

env-port-forward: ## env: Открыть порты сервисов окружения
	@docker compose up -d todo-port-forwarder

env-port-close: ## env: Закрыть порты сервисов окружения
	@docker compose down todo-port-forwarder

logs-cleanup: ## env: Очистить файлы логов из out/logs
	@read -p "Очистить все log файлы? Опасность утери логов. [y/N]: " ans; \
	if [ "$$ans" == "y" ]; then \
		rm -rf ${PROJECT_ROOT}/out/logs && \
		echo "Файлы логов очищены"; \
	else \
		echo "Очистка логов отменена"; \
	fi;

swagger-gen: ## env: Сгенерировать актуальную Swagger спецификацию
	@docker compose run --rm swagger \
		init \
		-g cmd/api/main.go \
		-o docs \
		--parseInternal \
		-parseDependency

ps: ## env: Посмотреть запущенные Docker Compose сервисы
	@docker compose ps

migrate-create: ## PostgreSQL: Создать новую версию схемы данных
	@if [ -z "$(seq)" ]; then \
		echo "Отсутствует необходимый параметр seq. Пример: make migrate-create seq=init"; \
		exit 1; \
	fi; \
	docker compose run --rm todo-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-up: ## PostgreSQL: Накатить миграции
	@make migrate-action action=up

migrate-down: ## PostgreSQL: Откатить миграции
	@make migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Отсутствует необходимый параметр action. Пример: make migrate-action action=up"; \
		exit 1; \
	fi; \
	docker compose run --rm todo-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@todo-postgres:5432/${POSTGRES_DB}?sslmode=disable \
		"$(action)"

dev: ## Golang приложение: Запустить локально на хост-системе (для локальной разработки)
	@export LOGGER_FOLDER=${PROJECT_ROOT}/out/logs && \
	export POSTGRES_HOST=127.0.0.1 && \
	export REDIS_HOST=127.0.0.1 && \
	export RMQ_HOST=127.0.0.1 && \
	go mod tidy && \
	go run ${PROJECT_ROOT}/cmd/api/main.go

help: ## Показать справку по командам
	@echo "=== Центр управления проектом ==="
	@echo ""
	@echo "Доступные команды:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

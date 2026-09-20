.PHONY: lint build run swag up down test-api lint-microservices build-microservices \
	k8s-deploy k8s-status k8s-ingress k8s-restart test-api-kubernetes help

ROOT := $(CURDIR)
MICROSERVICES := proxy events
GOLANGCI_LINT_IMAGE := golangci/golangci-lint:v2.6.2
K8S_DIR := src/kubernetes
K8S_NS := cinemaabyss

ifeq (lint,$(firstword $(MAKECMDGOALS)))
  SERVICE_GOALS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  $(eval $(SERVICE_GOALS):;@:)
endif

ifeq (build,$(firstword $(MAKECMDGOALS)))
  SERVICE_GOALS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  $(eval $(SERVICE_GOALS):;@:)
endif

ifeq (run,$(firstword $(MAKECMDGOALS)))
  RUN_TARGET := $(word 2,$(MAKECMDGOALS))
  $(eval $(RUN_TARGET):;@:)
endif

ifeq (swag,$(firstword $(MAKECMDGOALS)))
  SWAG_TARGET := $(word 2,$(MAKECMDGOALS))
  $(eval $(SWAG_TARGET):;@:)
endif

help:
	@echo "Usage:"
	@echo "  make lint <service> [service...]   Run golangci-lint (proxy, events, movies, monolith)"
	@echo "  make build <service> [service...]  Build service binary locally"
	@echo "  make run <service>                 Run service locally with default env"
	@echo "  make swag <service>              Regenerate Swagger docs (events)"
	@echo "  make lint-microservices            Lint proxy and events"
	@echo "  make build-microservices           Build proxy and events"
	@echo "  make up                            docker compose up -d --build"
	@echo "  make down                          docker compose down"
	@echo "  make test-api                      Run Postman tests (local environment)"
	@echo "  make k8s-deploy                    Deploy stack to minikube (ordered apply)"
	@echo "  make k8s-status                    kubectl get pod,svc,ingress in cinemaabyss"
	@echo "  make k8s-ingress                   Enable minikube ingress + apply ingress.yaml"
	@echo "  make k8s-restart                   Delete all pods (recreate with new images)"
	@echo "  make test-api-kubernetes           Postman tests against cinemaabyss.example.com"

lint:
	@test -n "$(SERVICE_GOALS)" || (echo "usage: make lint <service> [service...]"; exit 1)
	@for svc in $(SERVICE_GOALS); do \
		case $$svc in \
			proxy) dir=src/microservices/proxy ;; \
			events) dir=src/microservices/events ;; \
			movies) dir=src/microservices/movies ;; \
			monolith) dir=src/monolith ;; \
			*) echo "no such service: $$svc"; exit 1 ;; \
		esac; \
		test -d "$(ROOT)/$$dir" || (echo "no such service path: $$dir"; exit 1); \
		golangci-lint cache clean >/dev/null 2>&1 || true; \
		echo "lint $$svc ($$dir)"; \
		docker run --rm \
			--volume "$(ROOT)/$$dir:/app" \
			--workdir /app \
			$(GOLANGCI_LINT_IMAGE) \
			golangci-lint run || exit 1; \
	done

build:
	@test -n "$(SERVICE_GOALS)" || (echo "usage: make build <service> [service...]"; exit 1)
	@for svc in $(SERVICE_GOALS); do \
		case $$svc in \
			proxy) dir=src/microservices/proxy; bin=proxy-service; pkg=./cmd/api ;; \
			events) dir=src/microservices/events; bin=events-service; pkg=./cmd/api ;; \
			movies) dir=src/microservices/movies; bin=movies-service; pkg=. ;; \
			monolith) dir=src/monolith; bin=monolith; pkg=. ;; \
			*) echo "no such service: $$svc"; exit 1 ;; \
		esac; \
		echo "build $$svc ($$dir)"; \
		cd "$(ROOT)/$$dir" && mkdir -p bin && CGO_ENABLED=0 go build -o "bin/$$bin" "$$pkg"; \
	done

run:
	@test -n "$(RUN_TARGET)" || (echo "usage: make run <service>"; exit 1)
	@case $(RUN_TARGET) in \
		proxy) \
			cd "$(ROOT)/src/microservices/proxy" && \
			PORT=8000 \
			MONOLITH_URL=http://localhost:8080 \
			MOVIES_SERVICE_URL=http://localhost:8081 \
			EVENTS_SERVICE_URL=http://localhost:8082 \
			GRADUAL_MIGRATION=true \
			MOVIES_MIGRATION_PERCENT=50 \
			go run ./cmd/api ;; \
		events) \
			cd "$(ROOT)/src/microservices/events" && \
			PORT=8082 \
			KAFKA_BROKERS=localhost:9092 \
			go run ./cmd/api ;; \
		movies) \
			cd "$(ROOT)/src/microservices/movies" && \
			PORT=8081 \
			DB_CONNECTION_STRING="postgres://postgres:postgres_password@localhost:5432/cinemaabyss?sslmode=disable" \
			go run . ;; \
		monolith) \
			cd "$(ROOT)/src/monolith" && \
			PORT=8080 \
			DB_CONNECTION_STRING="postgres://postgres:postgres_password@localhost:5432/cinemaabyss?sslmode=disable" \
			go run . ;; \
		*) echo "no such service: $(RUN_TARGET)"; exit 1 ;; \
	esac

swag:
	@test -n "$(SWAG_TARGET)" || (echo "usage: make swag <service>"; exit 1)
	@case $(SWAG_TARGET) in \
		events) \
			cd "$(ROOT)/src/microservices/events" && \
			go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/api/main.go -o docs ;; \
		*) echo "swagger generation not configured for: $(SWAG_TARGET)"; exit 1 ;; \
	esac

lint-microservices:
	@$(MAKE) lint $(MICROSERVICES)

build-microservices:
	@$(MAKE) build $(MICROSERVICES)

up:
	docker compose up -d --build

down:
	docker compose down

test-api:
	cd tests/postman && npm run test:local

k8s-deploy:
	kubectl apply -f $(K8S_DIR)/namespace.yaml
	kubectl apply -f $(K8S_DIR)/configmap.yaml
	kubectl apply -f $(K8S_DIR)/secret.yaml
	kubectl apply -f $(K8S_DIR)/dockerconfigsecret.yaml
	kubectl apply -f $(K8S_DIR)/postgres-init-configmap.yaml
	kubectl apply -f $(K8S_DIR)/postgres.yaml
	@echo "Waiting for postgres..."
	kubectl -n $(K8S_NS) wait --for=condition=ready pod/postgres-0 --timeout=180s
	kubectl apply -f $(K8S_DIR)/kafka/kafka.yaml
	@echo "Waiting for kafka and zookeeper..."
	kubectl -n $(K8S_NS) wait --for=condition=ready pod/zookeeper-0 --timeout=300s
	kubectl -n $(K8S_NS) wait --for=condition=ready pod/kafka-0 --timeout=300s
	kubectl apply -f $(K8S_DIR)/monolith.yaml
	kubectl apply -f $(K8S_DIR)/movies-service.yaml
	kubectl apply -f $(K8S_DIR)/events-service.yaml
	kubectl apply -f $(K8S_DIR)/proxy-service.yaml

k8s-status:
	kubectl -n $(K8S_NS) get pod,svc,ingress

k8s-ingress:
	minikube addons enable ingress
	kubectl apply -f $(K8S_DIR)/ingress.yaml

k8s-restart:
	kubectl -n $(K8S_NS) delete pod --all

test-api-kubernetes:
	cd tests/postman && npm run test:kubernetes

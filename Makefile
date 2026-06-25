.PHONY: 
	build clean test run docker-build docker-up docker-down docker-logs 
	swagger-install swagger-gen swagger-clean test-validator test-health 
	test-health-unit docker-test-up docker-test-down docker-test-logs docker-test-build

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

build:
	$(GOBUILD) -o bin/api ./cmd/api

clean:
	$(GOCLEAN)
	rm -rf bin/

test:
	$(GOTEST) -v ./...

run:
	$(GOCMD) run ./cmd/api

tidy:
	$(GOMOD) tidy

docker-build:
	docker-compose build --no-cache=api -f docker-compose.yaml

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down -v

docker-logs:
	docker-compose logs -f api

swagger-install:
	@echo "Installing swagger..."
	go install github.com/swaggo/swag/cmd/swag@latest

swagger-gen:
	@echo "Generating swagger documentation..."
	swag init -g ./cmd/api/main.go -o docs --parseDependency --parseInternal

swagger-clean:
	@echo "Cleaning swagger docs..."
	rm -rf ./docs/

rebuild: clean tidy docker-build docker-up

swagger-setup: swagger-install swagger-gen

test-validator:
	go test -v ./internal/delivery/http/validator/password_validator_test.go

test-health:
	go test -v ./tests/integrations/ -run TestHealth

test-health-unit:
	go test -v ./internal/delivery/http/handler/health_handler_test.go

docker-test-up:
	docker-compose -f docker-compose-test.yaml up -d

docker-test-down:
	docker-compose -f docker-compose-test.yaml down -v

docker-test-logs:
	docker-compose -f docker-compose.test.yaml logs -f

docker-test-build:
	docker-compose -f docker-compose.test.yaml build --no-cache
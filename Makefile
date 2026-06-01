.PHONY: build clean test run docker-build docker-up docker-down docker-logs

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
	docker-compose build --no-cache=api

docker-up:
	docker-compose up -v

docker-down:
	docker-compose down -v

docker-logs:
	docker-compose logs -f api

rebuild: clean tidy docker-build docker-up
# Include variables from the .envrc file
include .envrc

# ======================================================================
# HELPERS
# ======================================================================

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

title ?=
page ?=
genre ?=
sort ?=
page_size ?=

## get-all: list all movies
.PHONY: get/all
get/all:
	@curl -s -G "localhost:4000/v1/movies" \
                $(if $(title),--data-urlencode "title=$(title)") \
                $(if $(genre),--data-urlencode "genres=$(genre)") \
                $(if $(page),--data-urlencode "page=$(page)") \
                $(if $(page_size),--data-urlencode "page_size=$(page_size)") \
                $(if $(sort),--data-urlencode "sort=$(sort)") | jq


.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]


# ======================================================================
# DEVELOPMENT
# ======================================================================

## run/api: run the cmd/api aplication
.PHONY: run/api
run/api:
	@go run ./cmd/api -db-dsn=${GREENLIGHT_DB_DSN}

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${GREENLIGHT_DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating new migration files form ${name}'
	goose -s create ${name} sql


## db/migrations/up: apply all up migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	goose up


# ======================================================================
# QUALITY CONTROL
# ======================================================================

## tidy: format all .go files and tidy module dependancies
.PHONY: tidy
tidy:
	@echo 'Tidying module dependencies...'
	go mod tidy	
	@echo 'Verifying and vendoring module dependencies...'
	go mod verify
	go mod vendor	
	@echo 'Formating .go files...'
	go fmt ./...


## audit: run quality control checks
.PHONY: audit
audit: 
	@echo 'Checking module dependencies'
	go mod tidy -diff
	go mod verify
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...


# ======================================================================
# BUILD
# ======================================================================


## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api

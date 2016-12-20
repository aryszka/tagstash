.PHONY: cpu mem

SOURCES = $(shell find . -name '*.go')
SQL_SOURCES = $(shell find . -name '*.sql')

PSQL_USER ?= tagstash
PSQL_DB ?= tagstash

default: build

gen: $(SQL_SOURCES)
	go generate

build: gen $(SOURCES)
	go build

install: $(SOURCES)
	go install

delete-postgres:
	psql --user $(PSQL_USER) -d $(PSQL_DB) -f sql/delete-db.sql

create-postgres:
	psql --user $(PSQL_USER) -d $(PSQL_DB) -f sql/create-db.sql

check: build
	go test -race

check-pq: build
	TEST_DB=postgres go test -race

shortcheck: build
	go test -test.short -run ^Test

shortcheck-pq: build
	TEST_DB=postgres go test -test.short -run ^Test

bench: build
	go test -cpuprofile cpu.out -memprofile mem.out -bench .

cpu:
	go tool pprof -top cpu.out # Run 'make bench' to generate profile.

mem:
	go tool pprof -top mem.out # Run 'make bench' to generate profile.

gencover: build
	go test -coverprofile cover.out

cover: gencover
	go tool cover -func cover.out

showcover: gencover
	go tool cover -html cover.out

fmt: $(SOURCES)
	gofmt -w -s ./*.go

vet: $(SOURCES)
	go vet

check-cyclo: $(SOURCES)
	gocyclo -over 12 .

check-ineffassign: $(SOURCES)
	ineffassign .

check-spell: $(SOURCES) README.md Makefile
	misspell -error README.md Makefile *.go

lint: $(SOURCES)
	golint -set_exit_status -min_confidence 0.9

precommit: build check fmt cover vet check-cyclo check-ineffassign check-spell lint
	# ok

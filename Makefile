.PHONY: build/api
build/api:
	@echo 'Building api...'
	go build -o=./bin/api ./cmd/api

.PHONY: run/bin/api
run/bin/api:
	@echo 'running api...'
	./bin/api



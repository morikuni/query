.PHONY: test
test:
	go test -race -v -count 1 ./...

.PHONY: coverage
coverage:
	go test -v -race -count 1 -covermode=atomic -coverprofile=coverage.out ./...
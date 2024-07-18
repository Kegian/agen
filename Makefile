build:
	go build -o out/agen ./cmd/agen/main.go

tidy:
	go mod tidy

fmt:
	gofmt -s -w .

lint:
	docker run -t --rm -v ${PWD}:/app -v ~/.cache/golangci-lint/v1.55.2:/root/.cache -w /app golangci/golangci-lint:v1.55.2 golangci-lint run -v

lint-clear-cache:
	rm -rf ~/.cache/golangci-lint/v1.55.2:/root/.cache

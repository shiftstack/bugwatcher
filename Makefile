lint:
	go fmt ./...

build:
	go build ./cmd/pretriage
	go build ./cmd/triage
	go build ./cmd/posttriage
	go build ./cmd/doctext

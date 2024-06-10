run-pretriage: pretriage
	./hack/run_with_env.sh ./$<
.PHONY: run-pretriage

lint:
	gofmt -w -s cmd pkg
.PHONY: lint

build: pretriage triage posttriage doctext

pretriage: cmd/pretriage pkg/query
	go build ./$<

triage: cmd/triage pkg/query
	go build ./$<

posttriage: cmd/posttriage pkg/query
	go build ./$<

doctext: cmd/doctext pkg/query
	go build ./$<

build: pretriage triage posttriage doctext

pretriage: cmd/pretriage pkg/jiraclient pkg/query pkg/slack pkg/team
	go build ./$<

triage: cmd/triage pkg/jiraclient pkg/query pkg/slack pkg/team
	go build ./$<

posttriage: cmd/posttriage pkg/jiraclient pkg/query
	go build ./$<

doctext: cmd/doctext pkg/jiraclient pkg/query pkg/slack pkg/team
	go build ./$<

lint:
	gofmt -w -s cmd pkg
.PHONY: lint

run-pretriage: pretriage
	./hack/run_with_env.sh ./$<
.PHONY: run-pretriage

run-triage: triage
	./hack/run_with_env.sh ./$<
.PHONY: run-triage

run-posttriage: posttriage
	./hack/run_with_env.sh ./$<
.PHONY: run-posttriage

run-doctext: doctext
	./hack/run_with_env.sh ./$<
.PHONY: run-doctext

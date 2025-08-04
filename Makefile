build:
	go build -o bin/gt-auto ./cmd/
run: build
	./bin/gt-auto --casePath pkg/testcase/testdata/risk_test_case.csv --config pkg/config/testdata/gw-auto-risk.toml
	./bin/gt-auto --casePath pkg/testcase/testdata/sse_test_case.csv --config pkg/config/testdata/gw-auto-sse.toml
	./bin/gt-auto --casePath pkg/testcase/testdata/szse_test_case.csv --config pkg/config/testdata/gw-auto-szse.toml
test:
	go test -v ./...
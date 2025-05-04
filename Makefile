build:
	go build -o bin/gt-auto ./cmd/
	go build -o bin/oms-client ./cmd/oms_client/
	go build -o bin/tgw-server ./cmd/tgw_server/
run: build
	./bin/gt-auto --casePath pkg/parser/testdata/test_case.csv
test:
	go test -v ./pkg/parser/...
	go test -v ./pkg/proto/...
	go test -v ./pkg/tcp/...
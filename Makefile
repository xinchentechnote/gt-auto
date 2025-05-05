build:
	go build -o bin/gt-auto ./cmd/
run: build
	./bin/gt-auto --casePath pkg/parser/testdata/test_case.csv
test:
	go test -v ./...
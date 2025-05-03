build:
	go build -o bin/gt-auto ./cmd/
run: build
	./bin/gt-auto
test:
	go test -v ./pkg/parser/...
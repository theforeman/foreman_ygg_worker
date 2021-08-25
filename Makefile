build: src/main.go src/server.go src/runner.go
	mkdir -p build
	CGO_ENABLED=0 go build -o build/foreman_worker src/main.go src/server.go src/runner.go

clean:
	rm -rf build

build: src/main.go src/server.go
	mkdir -p build
	go build -o build/foreman_worker src/main.go src/server.go

clean:
	rm -rf build

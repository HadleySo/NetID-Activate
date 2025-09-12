BINARY_NAME=NetIdActivate

build:
	go build -o ./bin/${BINARY_NAME} app.go

run:
	go build -o ./bin/${BINARY_NAME} app.go
	./bin/${BINARY_NAME}

clean:
	go clean
	rm ./bin/${BINARY_NAME}
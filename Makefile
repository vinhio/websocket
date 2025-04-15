.PHONY: run build certs

run:
	go run *.go

build:
	go build -o main *.go

certs:
	cd certs && sh certgen.sh && cd ../
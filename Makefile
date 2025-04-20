.PHONY: run build certs release

APP_NAME = app
BUILD_DIR = $(PWD)/build

run:
	go run *.go

build:
	go build -o main *.go

certs:
	cd certs && sh certgen.sh && cd ../

release:
	# 64-bit - Linux (amd64/arm64)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-amd64-linux *.go

	# App to server
	scp build/app-amd64-linux root@jivecode:/root/jiveim/tmp/
	# Restart server
	ssh root@jivecode 'cd /root/jiveim && ./restart.sh'

	# Clean to server
	ssh root@jivecode 'cd /root/jiveim/public && rm -fr * && rm -fr .*'
	# Public to server
	scp public/* root@jivecode:/root/jiveim/public/


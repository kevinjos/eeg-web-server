
NAME := eeg-server
ARCH := amd64
VERSION := 1.0
DATE := $(shell date)
COMMIT_ID := $(shell git rev-parse --short HEAD)
SDK_INFO := $(shell go version)
LD_FLAGS := '-X "main.buildInfo=Version: $(VERSION), commitID: $(COMMIT_ID), build date: $(DATE), SDK: $(SDK_INFO)"'

all: clean binaries jslibs

test:
	go test

binaries: test 
	GOOS=linux go build -ldflags $(LD_FLAGS) -o $(NAME)-linux-$(ARCH)

clean: 
	go clean
	rm -f eeg-server-linux-amd64

jslibs:
	mkdir -p js/libs
	mkdir -p js/build
ifeq ($(strip $(findstring d3.v3.min.js, $(wildcard js/libs/*.js))),)
	echo "no d3, calling wget"
	wget https://d3js.org/d3.v3.min.js -p -O js/libs/d3.v3.min.js
else
	@echo "d3 library exists, moving on"
endif
ifeq ($(strip $(findstring jquery.min.js, $(wildcard js/libs/*.js))),)
	echo "no jquery, calling wget"
	wget https://code.jquery.com/jquery-2.1.4.min.js -p -O js/libs/jquery.min.js
else
	@echo "jquery library exists, moving on"
endif
ifeq ($(strip $(findstring gl-plot3d, $(wildcard js/build/node_modules/*))),)
	echo "missing gl-plot3d node module for 3d plots, calling npm"
	npm install --prefix ./js/build gl-plot3d
else
	@echo "gl-plot3d node module exists, moving on"
endif
ifeq ($(strip $(findstring ndarray, $(wildcard js/build/node_modules/*))),)
	echo "missing ndarray node module for 3d plots, calling npm"
	npm install --prefix ./js/build ndarray
else
	@echo "ndarray node module exists, moving on"
endif
ifeq ($(strip $(findstring gl-surface3d, $(wildcard js/build/node_modules/*))),)
	echo "missing gl-surface3d node module for 3d plots, calling npm"
	npm install --prefix ./js/build gl-surface3d
else
	@echo "gl-surface3d node module exists, moving on"
endif
ifeq ($(strip $(findstring browserify, $(wildcard js/build/node_modules/*))),)
	echo "missing browserify node module for 3d plots, calling npm"
	npm install --prefix ./js/build browserify
else
	@echo "browserify node module exists, moving on"
endif
ifeq ($(strip $(findstring nodejs, $(wildcard /usr/bin/*))),nodejs)
	@echo "found nodejs binary, building bundle"
	/usr/bin/nodejs js/build/node_modules/browserify/bin/cmd.js js/3dplots.js -o js/libs/bundle.js
else ($(strip $(findstring node, $(wildcard /usr/bin/*))),node)
	@echo "found node binary, building bundle"
	js/build/node_modules/browserify/bin/cmd.js js/3dplots.js -o js/libs/bundle.js
else
	@echo "no node binary found, must install node"
endif

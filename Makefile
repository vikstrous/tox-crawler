MAKEFILE_DIR=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

all: run

build:
	docker build -t tox-crawler $(MAKEFILE_DIR)

run: build
	docker run -v $(MAKEFILE_DIR)/data:/go/src/github.com/vikstrous/tox-crawler/data tox-crawler

daemon: build
	docker run --name tox-crawler -d -p 80:7071 --restart always -v $(MAKEFILE_DIR)/data:/go/src/github.com/vikstrous/tox-crawler/data tox-crawler

publish:
	docker build -t vikstrous/tox-crawler $(MAKEFILE_DIR)
	docker push vikstrous/tox-crawler

.PHONY: all build run push

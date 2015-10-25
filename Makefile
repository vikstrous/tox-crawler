MAKEFILE_DIR=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

all: run

build:
	docker build -t tox-crawler $(MAKEFILE_DIR)

ifeq ($(BIND_ASSETS), 1)
ASSETS_MOUNT=-v $(MAKEFILE_DIR)/server:/go/src/github.com/vikstrous/tox-crawler/server
else
ASSETS_MOUNT=
endif

run: build
	docker run $(ASSETS_MOUNT) -v $(MAKEFILE_DIR)/data:/go/src/github.com/vikstrous/tox-crawler/data -p 80:7071 tox-crawler

daemon: build
	docker run --name tox-crawler -d -p 80:7071 --restart always -v $(MAKEFILE_DIR)/data:/go/src/github.com/vikstrous/tox-crawler/data tox-crawler

publish:
	docker build -t vikstrous/tox-crawler $(MAKEFILE_DIR)
	docker push vikstrous/tox-crawler

.PHONY: all build run push

IMAGE_NAME=chazari/balance_checker
TAG=latest

.PHONY: build push

build:
	docker build -t $(IMAGE_NAME):$(TAG) .

push: build
	docker push $(IMAGE_NAME):$(TAG)

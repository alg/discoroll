.PHONY: build
build:
	docker build . -t discoroll:latest --platform amd64

.PHONY: deploy
deploy:
	docker tag discoroll:latest registry.fly.io/discoroll:latest
	flyctl auth docker
	docker push registry.fly.io/discoroll:latest
	flyctl deploy --local-only


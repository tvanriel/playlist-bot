.PHONY: docker
docker:
	docker buildx build --builder mybuilder --no-cache --push --platform linux/amd64,linux/arm64 --no-cache -t mitaka8/playlist-bot:latest .

.PHONY: publish
publish:
	docker push mitaka8/playlist-bot:latest

.PHONY: web-dev
web-dev:
	CompileDaemon -build "go build -o /tmp/playlist-bot ." -command "/tmp/playlist-bot web"

.PHONY: discord-dev
discord-dev:
	CompileDaemon -build "go build -o /tmp/playlist-bot-discord ." -command "/tmp/playlist-bot-discord discord"

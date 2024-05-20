PACKAGE=tgquotebot
BUILDER=docker run --rm -v ./:/usr/local/go/src/$(PACKAGE) -w /usr/local/go/src/$(PACKAGE) golang:1.22
DOCKER=docker compose -f ./docker/docker-compose.yaml

build: clean
	@$(BUILDER) go build -o ./docker/images/app/build/app ./

	@cp ./font.ttf ./docker/images/app/build/font.ttf
	@cp ./quote.html ./docker/images/app/build/quote.html
	@cp ./avatar-placeholder.png ./docker/images/app/build/avatar-placeholder.png

	cp ./override/font.ttf ./docker/images/app/build/font.ttf || true
	cp ./override/quote.html ./docker/images/app/build/quote.html || true
	cp ./override/avatar-placeholder.png ./docker/images/app/build/avatar-placeholder.png || true

	@echo "build done"

clean:
	@rm -rf ./docker/images/app/build/*

start: build
	@$(DOCKER) up -d --build --remove-orphans

restart: stop start

status:
	@$(DOCKER) ps -a

stop:
	@$(DOCKER) down -v --remove-orphans

logs:
	@$(DOCKER) logs -f $(service)

attach:
	@$(DOCKER) exec $(service) bash

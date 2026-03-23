COMPOSE = sudo docker compose

.PHONY: up down restart build rebuild logs grid clean prune

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

restart:
	$(COMPOSE) down
	$(COMPOSE) up -d

build:
	$(COMPOSE) build

rebuild:
	$(COMPOSE) down
	$(COMPOSE) build --no-cache
	$(COMPOSE) up -d

logs:
	$(COMPOSE) logs -f --tail=50

grid:
	tmux new-session \; \
		split-window -h \; \
		select-pane -t 0 \; split-window -v \; \
		select-pane -t 2 \; split-window -v \; \
		select-pane -t 0 \; send-keys "$(COMPOSE) logs -f postgres" C-m \; \
		select-pane -t 1 \; send-keys "$(COMPOSE) logs -f solr" C-m \; \
		select-pane -t 2 \; send-keys "$(COMPOSE) logs -f audio-api" C-m \; \
		select-pane -t 3 \; send-keys "$(COMPOSE) logs -f bluppi-api" C-m

clean:
	$(COMPOSE) down -v

prune:
	docker system prune -f

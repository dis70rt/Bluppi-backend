up:
	@./start.sh
	sudo docker compose up -d --build

down:
	sudo docker compose down

clean:
	rm -rf __pycache__ .*.pid *.log

status:
	docker ps -a --format "table {{.ID}}\t{{.Image}}\t{{.Status}}\t{{.Names}}"
	@echo "Redis status:"
	sudo systemctl status redis-server
	@echo "PostgreSQL status:"
	sudo systemctl status postgresql

logs:
	multitail -cS dockerhttp -s 2 -l "docker logs -f bluppi-api" -l "docker logs -f bluppi-ws1"
build:
	sudo docker compose build
up:
	sudo docker compose up -d

down:
	sudo docker compose down

clean:
	rm -rf __pycache__ .*.pid *.log

stats:
	docker stats

logs:
	multitail -cS dockerhttp -s 3 \
		-l "docker logs -f bluppi-api" \
		-l "docker logs -f bluppi-ws1" \
		-l "docker logs -f bluppi-grpc"
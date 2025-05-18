APP_NAME = main:app
PORT = 8000
VENV = .venv

venv:
	python3 -m venv $(VENV)
	$(VENV)/bin/pip install --upgrade pip
	$(VENV)/bin/uv venv

install:
	$(VENV)/bin/uv install

up:
	sudo systemctl start redis-server
	sudo systemctl start postgresql
	source $(VENV)/bin/activate | $(VENV)/bin/uvicorn $(APP_NAME) --port $(PORT) --reload

up-detached:
	$(VENV)/bin/uvicorn $(APP_NAME) --port $(PORT) --reload &

down:
	pkill -f 'uvicorn'

clean:
	rm -rf $(VENV)
	rm -rf __pycache__

rebuild: clean venv install

status:
	ps aux | grep 'uvicorn'

logs:
	tail -f /var/log/uvicorn.log

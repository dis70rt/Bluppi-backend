APP_NAME = main:app
SOCKET_NAME = chat.server:app
SOCKET_PORT = 8080
PORT = 8000
VENV = .venv

PYTHON = $(VENV)/bin/python
UVICORN = $(VENV)/bin/uvicorn
PIP = $(VENV)/bin/pip

venv:
	python3 -m venv $(VENV)
	$(PIP) install --upgrade pip
	$(VENV)/bin/uv venv

install:
	$(VENV)/bin/uv install

start-services:
	sudo systemctl start redis-server
	sudo systemctl start postgresql

up: start-services
	@echo "Starting main app on port $(PORT)"
	$(UVICORN) $(APP_NAME) --port $(PORT) --reload &
	echo $$! > .main_pid
	@echo "Starting socket server on port $(SOCKET_PORT)"
	$(UVICORN) $(SOCKET_NAME) --port $(SOCKET_PORT) --reload &
	echo $$! > .socket_pid
	@echo "Both servers started."

up-detached: start-services
	nohup $(UVICORN) $(APP_NAME) --port $(PORT) --reload > app.log 2>&1 &
	echo $$! > .main_pid
	cd chat && nohup $(UVICORN) $(SOCKET_NAME) --port $(SOCKET_PORT) --reload > socket.log 2>&1 &
	echo $$! > .socket_pid

down:
	@echo "Stopping servers..."
	-kill -9 $$(cat .main_pid) 2>/dev/null || true
	-kill -9 $$(cat .socket_pid) 2>/dev/null || true
	rm -f .main_pid .socket_pid

clean:
	rm -rf $(VENV) __pycache__ .main_pid .socket_pid *.log

rebuild: clean venv install

status:
	ps aux | grep 'uvicorn'

logs:
	tail -f app.log socket.log

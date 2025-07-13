FROM python:3.12-slim

COPY --from=ghcr.io/astral-sh/uv:0.7.12 /uv /uvx /bin/

WORKDIR /workspace
COPY ./pyproject.toml .

RUN uv sync

COPY app/ ./app/
COPY chat/ ./chat/
COPY party/ ./party/
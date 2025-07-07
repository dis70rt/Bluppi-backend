FROM python:3.12-slim

COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

ENV VIRTUAL_ENV=/opt/.venv
ENV PATH="$VIRTUAL_ENV/bin:$PATH"

WORKDIR /workspace

COPY pyproject.toml uv.lock ./

RUN uv venv $VIRTUAL_ENV \
    && uv sync --frozen --no-cache \
    && uv pip install uvicorn

COPY app/ ./app/
COPY chat/ ./chat/

EXPOSE 8000 8080
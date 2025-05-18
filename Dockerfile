FROM python:3.12-slim

COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

WORKDIR /app

ENV PATH="/bin:$PATH"
COPY pyproject.toml uv.lock ./

RUN uv venv && \
    uv pip install --upgrade pip && \
    uv sync --frozen --no-cache

ENV PATH="/app/.venv/bin:$PATH"

EXPOSE 8000

CMD ["uvicorn", "main:app", "--reload"]

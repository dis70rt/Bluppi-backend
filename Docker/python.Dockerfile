FROM python:3.12-alpine

WORKDIR /app

RUN apk add --no-cache gcc musl-dev linux-headers

COPY searchMusic/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY searchMusic/ .

EXPOSE 50052
CMD ["python", "server.py"]

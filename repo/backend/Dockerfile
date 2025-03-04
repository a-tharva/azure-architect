FROM python:3.9-slim

WORKDIR /app
COPY . /app
RUN pip install flask grpcio grpcio-tools redis
EXPOSE 8000

CMD ["python", "app.py"]

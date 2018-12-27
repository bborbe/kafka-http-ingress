# Kafka HTTP Ingress

Send HTTP request to a topic.

## Run

```bash
go run main.go \
-port=8080 \
-kafka-brokers=kafka:9092 \
-kafka-topic=mytopic \
-v=2
```

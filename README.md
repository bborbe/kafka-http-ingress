# Kafka HTTP Ingress

Send HTTP request to a topic.

## Run

```bash
go run main.go \
-port=8080 \
-kafka-brokers=kafka:9092 \
-kafka-topic=http-requests \
-kafka-schema-registry-url=http://schema-registry:8081 \
-body-size-limit=1000000 \
-v=2
```

## Consume

```bash
docker exec -ti \
schema-registry \
kafka-avro-console-consumer \
--bootstrap-server=kafka:9092 \
--topic=http-requests \
--property=schema.registry.url=http://schema-registry:8081
```

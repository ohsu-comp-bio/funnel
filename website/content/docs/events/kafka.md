---
title: Kafka
menu:
  main:
    parent: Events
---

# Kafka

Funnel supports writing task events to a Kafka topic. To use this, add an event
writer to the config:

```
EventWriters:
  - kafka
  - log
  - rpc

Kafka:
  Servers:
    - localhost:9092
  Topic: funnel-events
```

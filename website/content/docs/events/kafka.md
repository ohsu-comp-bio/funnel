---
title: Kafka
menu:
  main:
    parent: Events
---

# Kafka

Funnel supports writing task events to a Kafka topic. To use this, add an event
writer to the worker config:

```
Worker:
  ActiveEventWriters:
    - kafka
    - log
    - rpc
  EventWriters:
    Kafka:
      Servers:
        - localhost:9092
      Topic: funnel-events
```

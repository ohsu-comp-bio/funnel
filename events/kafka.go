package events

import (
	"github.com/Shopify/sarama"
	"github.com/ohsu-comp-bio/funnel/config"
)

// KafkaWriter writes events to a Kafka topic.
type KafkaWriter struct {
	conf     config.Kafka
	producer sarama.SyncProducer
}

// NewKafkaWriter creates a new event writer for writing events to a Kafka topic.
func NewKafkaWriter(conf config.Kafka) (*KafkaWriter, error) {
	producer, err := sarama.NewSyncProducer(conf.Servers, nil)
	if err != nil {
		return nil, err
	}
	return &KafkaWriter{conf, producer}, nil
}

// Close closes the Kafka producer, cleaning up resources.
func (k *KafkaWriter) Close() error {
	return k.producer.Close()
}

// Write writes the event. Events may be sent in batches in the background by the
// Kafka client library. Currently stdout, stderr, and system log events are dropped.
func (k *KafkaWriter) Write(ev *Event) error {

	switch ev.Type {
	case Type_EXECUTOR_STDOUT, Type_EXECUTOR_STDERR, Type_SYSTEM_LOG:
		return nil
	}

	s, err := Marshal(ev)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: k.conf.Topic,
		Key:   nil,
		Value: sarama.StringEncoder(s),
	}
	_, _, err = k.producer.SendMessage(msg)
	return err
}

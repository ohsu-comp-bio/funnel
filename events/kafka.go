package events

import (
	"github.com/Shopify/sarama"
	"github.com/ohsu-comp-bio/funnel/config"
)

type KafkaWriter struct {
	conf     config.Kafka
	producer sarama.SyncProducer
}

func NewKafkaWriter(conf config.Kafka) (*KafkaWriter, error) {
	producer, err := sarama.NewSyncProducer(conf.Servers, nil)
	if err != nil {
		return nil, err
	}
	return &KafkaWriter{conf, producer}, nil
}

func (k *KafkaWriter) Close() error {
	return k.producer.Close()
}

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

package kafka

import (
	"context"
	"fmt"
	"log"

	kafkago "github.com/segmentio/kafka-go"
)

type Consumer struct {
	readers []*kafkago.Reader
}

func NewConsumer(brokers string, groupID string) *Consumer {
	if groupID == "" {
		groupID = defaultConsumerGroup
	}

	brokerList := parseBrokers(brokers)
	topics := Topics()
	readers := make([]*kafkago.Reader, 0, len(topics))

	for _, topic := range topics {
		readers = append(readers, kafkago.NewReader(kafkago.ReaderConfig{
			Brokers: brokerList,
			GroupID: groupID,
			Topic:   topic,
		}))
	}

	return &Consumer{readers: readers}
}

func (c *Consumer) Start(ctx context.Context) {
	for _, reader := range c.readers {
		go c.consume(ctx, reader)
	}
}

func (c *Consumer) Close() error {
	var firstErr error
	for _, reader := range c.readers {
		if err := reader.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("close reader for topic %s: %w", reader.Config().Topic, err)
		}
	}

	return firstErr
}

func (c *Consumer) consume(ctx context.Context, reader *kafkago.Reader) {
	topic := reader.Config().Topic

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}

			log.Printf("kafka consumer read error: topic=%s err=%v", topic, err)

			continue
		}

		log.Printf(
			"processed event: topic=%s partition=%d offset=%d payload=%s",
			msg.Topic,
			msg.Partition,
			msg.Offset,
			string(msg.Value),
		)
	}
}

package kafka

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	kafkago "github.com/segmentio/kafka-go"
)

const defaultConsumerGroup = "events-service"

var ErrNoBrokers = errors.New("no kafka brokers configured")

type PublishResult struct {
	Partition int
	Offset    int64
}

type Producer struct {
	writers map[string]*kafkago.Writer
	brokers []string
	mu      sync.Mutex
}

func NewProducer(brokers string) *Producer {
	return &Producer{
		brokers: parseBrokers(brokers),
		writers: make(map[string]*kafkago.Writer),
	}
}

func (p *Producer) Publish(ctx context.Context, topic string, value []byte) (PublishResult, error) {
	writer := p.writer(topic)

	msg := kafkago.Message{
		Value: value,
	}

	if err := writer.WriteMessages(ctx, msg); err != nil {
		return PublishResult{}, fmt.Errorf("publish message to topic %s: %w", topic, err)
	}

	return PublishResult{
		Partition: msg.Partition,
		Offset:    msg.Offset,
	}, nil
}

func (p *Producer) Ping(ctx context.Context) error {
	if len(p.brokers) == 0 {
		return ErrNoBrokers
	}

	var firstErr error

	for _, broker := range p.brokers {
		err := pingBroker(ctx, broker)
		if err == nil {
			return nil
		}

		if firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func pingBroker(ctx context.Context, broker string) error {
	conn, err := kafkago.DialContext(ctx, "tcp", broker)
	if err != nil {
		return fmt.Errorf("dial kafka broker %s: %w", broker, err)
	}

	_, err = conn.Brokers()
	if closeErr := conn.Close(); closeErr != nil {
		log.Printf("close kafka ping connection %s: %v", broker, closeErr)
	}

	if err != nil {
		return fmt.Errorf("list kafka brokers via %s: %w", broker, err)
	}

	return nil
}

func (p *Producer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var firstErr error
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("close writer for topic %s: %w", topic, err)
		}
	}

	p.writers = make(map[string]*kafkago.Writer)

	return firstErr
}

func (p *Producer) writer(topic string) *kafkago.Writer {
	p.mu.Lock()
	defer p.mu.Unlock()

	if writer, ok := p.writers[topic]; ok {
		return writer
	}

	writer := &kafkago.Writer{
		Addr:         kafkago.TCP(p.brokers...),
		Topic:        topic,
		Balancer:     &kafkago.LeastBytes{},
		RequiredAcks: kafkago.RequireOne,
	}

	p.writers[topic] = writer

	return writer
}

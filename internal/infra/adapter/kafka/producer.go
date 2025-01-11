package kafka

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

type WriterInterface interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}
type DynamicProducer struct {
	messages []kafka.Message
	mu       sync.Mutex
	writer   WriterInterface
	wg       sync.WaitGroup
	stopChan chan struct{}
}

func NewDynamicKafkaProducer(brokerAddress, topic string) *DynamicProducer {
	err := createTopic(brokerAddress, topic, 20, 1)
	if err != nil {
		panic(err)
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{brokerAddress},
		Topic:   topic,
	})

	producer := &DynamicProducer{
		writer:   writer,
		stopChan: make(chan struct{}),
	}

	producer.startWorker()

	return producer
}

func (p *DynamicProducer) Produce(key string, value []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.messages = append(p.messages, kafka.Message{
		Key:   []byte(key),
		Value: value,
	})

	return nil
}

func (p *DynamicProducer) startWorker() {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for {
			select {
			case <-p.stopChan:
				return
			default:
				p.ProcessQueue()
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func (p *DynamicProducer) ProcessQueue() {
	p.mu.Lock()
	if len(p.messages) == 0 {
		p.mu.Unlock()
		return
	}

	message := p.messages[0]
	p.messages = p.messages[1:]
	p.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		log.Printf("Erro ao enviar mensagem para o Kafka: %v", err)
	}
}

func createTopic(brokerAddress, topic string, partitions, replicationFactor int) error {
	conn, err := kafka.Dial("tcp", brokerAddress)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao Kafka broker: %w", err)
	}

	defer func(conn *kafka.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	partitionsList, err := conn.ReadPartitions(topic)
	if err == nil && len(partitionsList) > 0 {
		log.Printf("O tópico '%s' já existe.", topic)

		return nil
	}

	err = conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     partitions,
		ReplicationFactor: replicationFactor,
	})
	if err != nil {
		return fmt.Errorf("erro ao criar tópico '%s': %w", topic, err)
	}

	log.Printf("Tópico '%s' criado com sucesso.", topic)

	return nil
}

func (p *DynamicProducer) Close() {
	close(p.stopChan)
	p.wg.Wait()

	p.mu.Lock()
	defer p.mu.Unlock()

	for len(p.messages) > 0 {
		p.ProcessQueue()
	}

	if err := p.writer.Close(); err != nil {
		log.Printf("Erro ao fechar writer Kafka: %v", err)

		return
	}
	log.Println("Writer Kafka fechado com sucesso")
}

func WaitForKafka(brokerAddress string, timeout time.Duration) {
	start := time.Now()
	for {
		conn, err := net.DialTimeout("tcp", brokerAddress, 5*time.Second)
		if err == nil {
			err := conn.Close()
			if err != nil {
				return
			}

			log.Printf("Conexão com Kafka bem-sucedida após %v", time.Since(start))

			return
		}

		if time.Since(start) > timeout {
			log.Fatalf("Timeout ao tentar conectar ao Kafka: %v", err)
		}

		log.Println("Kafka ainda não está disponível, tentando novamente em 5s...")
		time.Sleep(5 * time.Second)
	}
}

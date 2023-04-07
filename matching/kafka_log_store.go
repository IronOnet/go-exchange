package matching

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaLogStore struct {
	logWriter *kafka.Writer 
}

func NewKafkaLogStore(productId string, brokers []string) *KafkaLogStore{
	s := &KafkaLogStore{} 

	s.logWriter = kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers, 
		Topic: topicBookMessagePrefix + productId, 
		Balancer: &kafka.LeastBytes{}, 
		BatchTimeout: 5 * time.Millisecond,
	})

	return s 
}

func (s *KafkaLogStore) Store(logs []interface{}) error{
	var messages []kafka.Message 
	for _, log := range logs{
		val, err := json.Marshal(log) 
		if err != nil{
			return err 
		}

		messages = append(messages, kafka.Message{Value: val}) 

	}

	return s.logWriter.WriteMessages(context.Background(), messages...)
}
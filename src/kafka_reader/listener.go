package kafka_reader

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"net/http"

	kafka "github.com/segmentio/kafka-go"
)

type KafkaReader struct {
	Config kafka.ReaderConfig
	Reader kafka.Reader
	Host   string
	Topic  string
	url    string
}

func InitKafkaReader(host string, topic string, partition int) (KafkaReader, error) {
	config := kafka.ReaderConfig{
		Brokers:  []string{host},
		GroupID:  "test-consumer-group",
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	}

	reader := kafka.NewReader(config)

	url := "http://srv1c.itgrn/UNF_DEMO_IND/hs/kafka/document"
	KafkaReader := KafkaReader{config, *reader, host, topic, url}

	return KafkaReader, nil
}

func (kr *KafkaReader) Read() {
	log.Printf("Start kafka reader. Host: %s Topic: %s", kr.Host, kr.Topic)
	for {
		m, err := kr.Reader.ReadMessage(context.Background())
		if err != nil {
			break
		}
		log.Printf("message at topic/partition/offset %v/%v/%v: %s\n", m.Topic, m.Partition, m.Offset, string(m.Value))
		req, err := http.NewRequest("POST", kr.url, bytes.NewBuffer([]byte(m.Value)))
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		log.Println("response Status:", resp.Status)
		log.Println("response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("response Body:", string(body))
	}
	log.Printf("Stop kafka reader.")
}

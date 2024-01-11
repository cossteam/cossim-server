package msg_queue

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type RabbitMQ struct {
	connection *amqp.Connection
	//channel    *amqp.Channel
}

// NewRabbitMQ 用于创建 RabbitMQ 结构实例
func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	//ch, err := conn.Channel()
	//if err != nil {
	//	conn.Close()
	//	return nil, err
	//}

	return &RabbitMQ{
		connection: conn,
		//channel:    ch,
	}, nil
}

// Close 关闭 RabbitMQ 连接和通道
func (r *RabbitMQ) Close() {
	//r.channel.Close()
	r.connection.Close()
}

// PublishMessage 用于发布消息
func (r *RabbitMQ) PublishMessage(queueName string, body interface{}) error {
	channel, err := r.connection.Channel()
	if err != nil {
		return err
	}
	q, err := channel.QueueDeclare(
		queueName,
		false,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println("Failed to declare a queue :", err)
		return err
	}
	var msg []byte
	if body != nil {
		//解析成json
		msg, err = json.Marshal(&body)
		if err != nil {
			return err
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	fmt.Println("发布消息：", string(msg))
	return channel.PublishWithContext(ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		})
}

// ConsumeMessages 用于消费消息
func (r *RabbitMQ) ConsumeMessages(queueName string) (<-chan amqp.Delivery, error) {
	channel, err := r.connection.Channel()
	if err != nil {
		return nil, err
	}
	q, err := channel.QueueDeclare(
		queueName,
		false,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
}

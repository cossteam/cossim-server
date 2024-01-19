package msg_queue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cossim/coss-server/pkg/encryption"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

// NewRabbitMQ 用于创建 RabbitMQ 结构实例
func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &RabbitMQ{
		connection: conn,
		channel:    ch,
	}, nil
}
func (r *RabbitMQ) NewChannel() (*amqp.Channel, error) {
	ch, err := r.connection.Channel()
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func (r *RabbitMQ) GetChannel() *amqp.Channel {
	return r.channel
}

func (r *RabbitMQ) GetConnection() *amqp.Connection {
	return r.connection
}

// Close 关闭 RabbitMQ 连接和通道
func (r *RabbitMQ) Close() {
	r.channel.Close()
	r.connection.Close()
}

// PublishMessage 用于发布消息
func (r *RabbitMQ) PublishMessage(queueName string, body interface{}) error {
	q, err := r.channel.QueueDeclare(
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
	return r.channel.PublishWithContext(context.Background(),
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		})
}

func (r *RabbitMQ) PublishEncryptedMessage(queueName string, body string) error {
	q, err := r.channel.QueueDeclare(
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
	var enMsg encryption.SecretResponse

	var msg []byte
	if body != "" {
		//解析成json
		err = json.Unmarshal([]byte(body), &enMsg)
		if err != nil {
			return err
		}
		msg, err = json.Marshal(enMsg)
		if err != nil {
			return err
		}
	}
	return r.channel.PublishWithContext(context.Background(),
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msg,
		})
}

// ConsumeMessages 用于消费消息
func ConsumeMessages(queueName string, channel *amqp.Channel) (amqp.Delivery, bool, error) {
	q, err := channel.QueueDeclare(
		queueName,
		false,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return amqp.Delivery{}, false, err
	}
	msg, ok, err := channel.Get(q.Name, true)
	if err != nil {
		fmt.Println(err)
		return amqp.Delivery{}, false, err
	}

	return msg, ok, nil
}

func (r *RabbitMQ) ConsumeMessagesWithChan(queueName string) (<-chan amqp.Delivery, error) {
	//channel, err := r.NewChannel()
	q, err := r.channel.QueueDeclare(
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
	return r.channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
}

// 删除空闲的队列和资源
func (r *RabbitMQ) DeleteEmptyQueue(queueName string) error {
	//channel, err := r.NewChannel()
	_, err := r.channel.QueueDelete(queueName, true, false, false)
	if err != nil {
		return err
	}
	return nil
}

func (r *RabbitMQ) ConsumeServiceMessages(queueName ServiceType, exchangeName string) (<-chan amqp.Delivery, error) {
	//channel, err := r.NewChannel()
	err := r.channel.ExchangeDeclare(
		exchangeName,
		amqp.ExchangeDirect,
		false,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	q, err := r.channel.QueueDeclare(
		string(queueName),
		false,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	//绑定队列到交换机
	err = r.channel.QueueBind(q.Name, "", Service_Exchange, false, nil)
	if err != nil {
		return nil, err
	}
	return r.channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
}
func (r *RabbitMQ) PublishServiceMessage(serviceName, targetName ServiceType, exchangeName string, action ServiceActionType, body interface{}) error {
	//channel, err := r.NewChannel()
	//if err != nil {
	//	return err
	//}
	err := r.channel.ExchangeDeclare(
		exchangeName,
		amqp.ExchangeDirect,
		false,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	q, err := r.channel.QueueDeclare(
		string(targetName),
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

	//绑定队列到交换机
	err = r.channel.QueueBind(q.Name, string(targetName), Service_Exchange, false, nil)
	if err != nil {
		return err
	}
	var msg []byte
	data := &ServiceQueueMsg{
		Action: action,
		Data:   body,
		Form:   serviceName,
	}
	if body != nil {
		//解析成json
		msg, err = json.Marshal(&data)
		if err != nil {
			return err
		}
	}
	return r.channel.PublishWithContext(context.Background(),
		Service_Exchange,
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		})
}

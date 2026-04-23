package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ride-sharing/shared/contracts"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	TripExchange = "trip"
)

type RabbitMQ struct {
	conn    *amqp.Connection // 一个TCP连接到RABBIT MQ服务器
	Channel *amqp.Channel    // 代表一个会话/通道，复用上面的TCP连接，但是可随时关闭创建，更轻量
}

func NewRabbitMQ(uri string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to rabbitmq: %v", err)
	}

	// 打开通道
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Failed to open chennel: %v", err)
	}

	rmq := &RabbitMQ{
		conn:    conn,
		Channel: ch,
	}

	if err := rmq.setupExchangesAndQueues(); err != nil {
		return nil, fmt.Errorf("Failed to set up exchanges and queues: %v", err)
	}

	return rmq, nil
}

type MessageHandler func(context.Context, amqp.Delivery) error

func (rmq *RabbitMQ) ConsumeMessages(queueName string, handler MessageHandler) error {

	err := rmq.Channel.Qos(
		1,     // 预取数量，告诉服务器 在收到我这个消费者的任何确认之前，最多可以一次性发给我多少条消息（我本地最多可以缓存多少条数据）
		0,     // 预取大小，在我确认之前，你发送给我的所有未确认消息的总大小不要超过X字节
		false, // 全局标志，如果是TRUE，设置就应用于同一个TCP连接上的所有通道channel了，比如说如果预取数量等于10，且一共3个通道，那么3个通道共享10个额度。
		// 如果是FALSE，则只应用于当前这个通道
	)

	if err != nil {
		return fmt.Errorf("failed to set Qos: %v", err)
	}

	msgs, err := rmq.Channel.Consume( // 这里的CONSUME方法返回的是一个通道，不是批量导出，所以下面用FOR循环监听
		queueName, // queue
		"",        // consumer 消费者标识，空字符串 = 自动生成；作用就是后续可以取消特定的消费者
		false,     // auto_ack 自动确认；如果设置为TRUE,那么RABBITMQ发送消息后就标记为已确认，如果消费者崩溃，消息就会永久丢失；如果
		// 设置为FALSE，则需要手动调用msg.Ack(false)，这样可以确保消息不丢失
		false, // exclusive 是否标识为只有当前这个连接可以消费此队列，如果设置为TRUE，独占，其他连接尝试消费就会报错
		false, // no-local 本地标志，如果设置为true，不接受本连接发布的消息
		false, // no-wait 非等待模式，如果设false，会等待服务器确认，确保消费者注册成功；反之，不等待确认，可能因队列不存在等原因失败，且不报错
		nil,   // args
	)

	if err != nil {
		return err
	}

	ctx := context.Background()

	go func() {
		for msg := range msgs {
			log.Printf("Received a message: %s", msg.Body)

			if err := handler(ctx, msg); err != nil {
				log.Printf("Failed to handle the message: %v", err)
				if nackErr := msg.Nack(false, false); nackErr != nil {
					log.Printf("Failed to Nack the message: %v", nackErr)
				}
			}

			// 确认消息当handler处理正常
			if ackErr := msg.Ack(false); ackErr != nil {
				log.Printf("Failed to ack the message: %v", ackErr)
			}
		}
	}()

	return nil
}

func (rmq *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, message contracts.AmqpMessage) error {
	// 对message 整体再做一次JSON序列化
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return nil
	}

	return rmq.Channel.PublishWithContext(ctx,
		TripExchange, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediately
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         jsonMessage,
			DeliveryMode: amqp.Persistent,
		},
	)
}

func (rmq *RabbitMQ) setupExchangesAndQueues() error {
	err := rmq.Channel.ExchangeDeclare(
		TripExchange, // 名称
		"Topic",      // 类型，有fanout, direct
		true,         // 是否持久化
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to declare exchange: %v", err)
	}

	if err := rmq.declareAndBindQueue(
		FindAvailableDriversQueue,
		[]string{contracts.TripEventCreated, contracts.TripEventDriverNotInterested},
		TripExchange,
	); err != nil {
		return err
	}

	return nil
}

func (rmq *RabbitMQ) declareAndBindQueue(queueName string, messageType []string, exchange string) error {
	// 创建QUEUE
	q, err := rmq.Channel.QueueDeclare(
		queueName, // name
		true,      // durable 是否持久化，如果为TRUE，则会保存在磁盘中
		false,     // delete when unused 是否自动删除
		false,     // exclusive
		false,     // no-wait是否异步
		nil,
	)

	if err != nil {
		log.Fatalf("Failed to declare queues: %v", err)
	}

	// 绑定队列几路由键
	for _, msg := range messageType {
		err := rmq.Channel.QueueBind(
			q.Name,   // queue name
			msg,      // routing key
			exchange, // exchange
			false,    //
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue to %s: %v", queueName, err)
		}
	}

	return nil
}

func (rmq *RabbitMQ) Close() {
	if rmq.conn != nil {
		rmq.conn.Close()
	}
	if rmq.Channel != nil {
		rmq.Channel.Close()
	}
}

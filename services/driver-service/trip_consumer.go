package main

import (
	"context"
	"encoding/json"
	"log"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"

	pb "ride-sharing/shared/proto/trip"

	"github.com/rabbitmq/amqp091-go"
)

type TripConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  *Service
}

func NewTripConsumer(rmq *messaging.RabbitMQ, service *Service) *TripConsumer {
	return &TripConsumer{
		rabbitmq: rmq,
		service:  service,
	}
}

func (c *TripConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(messaging.FindAvailableDriversQueue, func(ctx context.Context, msg amqp091.Delivery) error {
		var tripEvent contracts.AmqpMessage
		if err := json.Unmarshal(msg.Body, &tripEvent); err != nil {
			log.Printf("Failed to unmarshall message: %v", err)
			return err
		}

		// 对AMQP MESSAGE的BODY做反序列化
		var body pb.Trip
		if err := json.Unmarshal(tripEvent.Data, &body); err != nil {
			return err
		}

		switch msg.RoutingKey {
		case contracts.TripEventCreated, contracts.TripEventDriverNotInterested:
			return c.handleFindAndNotifyDrivers(ctx, &body)
		}

		log.Printf("unknown trip event: %+v", msg)
		return nil
	})
}

func (c *TripConsumer) handleFindAndNotifyDrivers(ctx context.Context, msg *pb.Trip) error {
	suitableDrivers := c.service.FindAvailableDrivers(msg.Fare.PackageSlug)

	// 检查是否有合适的司机
	if suitableDrivers == nil {
		// 如果没有合适的司机，发出通知说没有司机接单
		if err := c.rabbitmq.PublishMessage(ctx, contracts.TripEventNoDriversFound, contracts.AmqpMessage{
			OwnerID: msg.UserID,
		}); err != nil {
			log.Printf("faild to publish message to exchange : %v", err)
		}

		return nil
	}

	suitableDriver := suitableDrivers[0]
	marshalledEvent, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// 给司机发送行程等信息
	if err := c.rabbitmq.PublishMessage(ctx, contracts.DriverCmdTripRequest, contracts.AmqpMessage{
		OwnerID: suitableDriver,
		Data:    marshalledEvent,
	}); err != nil {
		log.Printf("faild to publish message to exchange : %v", err)
	}

	return nil
}

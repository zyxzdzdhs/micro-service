package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"
	pb "ride-sharing/shared/proto/driver"
	pbTrip "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/retry"

	"github.com/rabbitmq/amqp091-go"
)

type driverConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  domain.TripService
}

func NewDriverConsumer(rmq *messaging.RabbitMQ, service domain.TripService) *driverConsumer {
	return &driverConsumer{
		rabbitmq: rmq,
		service:  service,
	}
}

func (drvConsumer *driverConsumer) Listen() error {
	return drvConsumer.rabbitmq.ConsumeMessages(messaging.DriverCmdTripResponseQueue, func(ctx context.Context, msg amqp091.Delivery) error {
		var message contracts.AmqpMessage
		if err := json.Unmarshal(msg.Body, &message); err != nil {
			log.Printf("Failed to unmarshall message: %v", err)
			return err
		}

		// 对AMQP MESSAGE的BODY做反序列化
		var body messaging.DriverTripResponseData
		if err := json.Unmarshal(message.Data, &body); err != nil {
			return err
		}

		switch msg.RoutingKey {
		case contracts.DriverCmdTripAccept:
			if err := drvConsumer.handleTripAccepted(ctx, body.TripID, body.Driver); err != nil {
				log.Printf("Failed to handle the trip accept: %v", err)
				return err
			}
		case contracts.DriverCmdTripDecline:
			if err := drvConsumer.handleTripDecline(ctx, body.TripID, body.RiderID); err != nil {
				log.Printf("Failed to handle the trip decline: %v", err)
				return err
			}
			return nil
		}

		log.Printf("unknown trip event: %+v", msg)
		return nil
	})
}

func (drvConsumer *driverConsumer) handleTripDecline(ctx context.Context, tripID string, riderID string) error {
	trip, err := drvConsumer.service.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}

	newPayload := trip.ToProto()

	marshalledPayload, err := json.Marshal(newPayload)
	if err != nil {
		return err
	}

	if err := drvConsumer.rabbitmq.PublishMessage(ctx, contracts.TripEventDriverNotInterested,
		contracts.AmqpMessage{
			OwnerID: riderID,
			Data:    marshalledPayload,
		},
	); err != nil {
		return err
	}

	return nil
}

func (drvConsumer *driverConsumer) handleTripAccepted(ctx context.Context, tripID string, driver *pb.Driver) error {
	// 获取行程
	trip, err := drvConsumer.service.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}

	if trip == nil {
		return fmt.Errorf("trip was not found %s", tripID)
	}

	// 更新行程
	if err := drvConsumer.service.UpdateTrip(ctx, tripID, "accepted", driver); err != nil {
		return err
	}

	// 再次获取行程
	trip, err = drvConsumer.service.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}

	// 司机被分配好了，下一步：发送到RABBIT MQ
	marshalledTrip, err := json.Marshal(trip)
	if err != nil {
		return err
	}

	if err := drvConsumer.rabbitmq.PublishMessage(ctx, contracts.TripEventDriverAssigned, contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    marshalledTrip,
	}); err != nil {
		return err
	}

	return nil
}

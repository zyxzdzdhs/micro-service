package events

import (
	"context"
	"encoding/json"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"
)

type TripEventPublisher struct {
	rabbitmq *messaging.RabbitMQ
}

func NewTripEventPublisher(rabbitmq *messaging.RabbitMQ) *TripEventPublisher {
	return &TripEventPublisher{
		rabbitmq: rabbitmq,
	}
}

func (pub *TripEventPublisher) PublishTripCreated(ctx context.Context, trip *domain.TripModel) error {
	tripEventJson, err := json.Marshal(trip.ToProto()) // 这里对tripmodel的操作是因为在消费者侧无法通过反序列化拿到TripModel,所以用PROTOBUF的PB.TRIP
	if err != nil {
		return err
	}

	return pub.rabbitmq.PublishMessage(ctx, contracts.TripEventCreated, contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    tripEventJson,
	})
}

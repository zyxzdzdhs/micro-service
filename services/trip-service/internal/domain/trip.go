package domain

import (
	"context"
	pbd "ride-sharing/shared/proto/driver"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TripModel struct {
	ID       primitive.ObjectID
	UserID   string
	Status   string
	RideFare *RideFareModel
	Driver   *pb.TripDriver
}

type Trip struct {
	ID       primitive.ObjectID
	UserID   string
	Status   string
	RideFare RideFareModel
}

func (t *TripModel) ToProto() *pb.Trip {
	return &pb.Trip{
		Id:     t.ID.Hex(),
		UserID: t.UserID,
		Fare:   t.RideFare.ToProto(),
		Status: t.Status,
		Driver: t.Driver,
		Route:  t.RideFare.Route.ToProto(),
	}
}

type TripRepository interface {
	CreateTrip(ctx context.Context, trip *TripModel) (*TripModel, error)
	SaveRideFare(ctx context.Context, f *RideFareModel) error

	GetRideFareByID(ctx context.Context, id string) (*RideFareModel, error)

	GetTripByID(ctx context.Context, id string) (*TripModel, error)
	UpdateTrip(ctx context.Context, tripID string, status string, driver *pbd.Driver) error
}

type TripService interface {
	CreateTrip(ctx context.Context, fare *RideFareModel) (*TripModel, error)
	GetRoute(ctx context.Context, pickUp, destination *types.Coordinate) (*types.OsrmApiResource, error)
	EstimatePackagesPriceWithRoute(route *types.OsrmApiResource) []*RideFareModel
	GenerateTripFares(ctx context.Context, fares []*RideFareModel, userID string, route *types.OsrmApiResource) ([]*RideFareModel, error)

	GetAndValidateFare(ctx context.Context, fareID, userID string) (*RideFareModel, error)

	GetTripByID(ctx context.Context, id string) (*TripModel, error)
	UpdateTrip(ctx context.Context, tripID string, status string, driver *pbd.Driver) error
}

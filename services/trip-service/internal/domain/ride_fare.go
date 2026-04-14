package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"
)

type RideFareModel struct {
	ID                primitive.ObjectID
	UserID            string
	PackageSlug       string
	TotalPriceInCents float64
	Route             types.OsrmApiResource
}

func (f *RideFareModel) ToProto() *pb.RideFare {
	return &pb.RideFare{
		Id:              f.ID.Hex(),
		UserID:          f.UserID,
		PackageSlug:     f.PackageSlug,
		TotalPriceCents: f.TotalPriceInCents,
	}
}

func ToRideFareProto(fares []*RideFareModel) []*pb.RideFare {
	result := make([]*pb.RideFare, len(fares))
	for i, fare := range fares {
		result[i] = fare.ToProto()
	}
	return result
}

package main

import (
	"context"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
)

func main() {
	ctx := context.Background()
	fare := &domain.RideFareModel{
		UserID: "42",
	}
	memRepo := repository.NewInmemRepository()
	svc := service.NewService(memRepo)
	svc.CreateTrip(ctx, fare)
}

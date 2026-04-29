package messaging

import (
	pb "ride-sharing/shared/proto/driver"
)

const (
	FindAvailableDriversQueue  = "find_available_drivers"
	DriverCmdTripRequestQueue  = "driver_cmd_trip_request"
	DriverCmdTripResponseQueue = "driver_cmd_trip_response"

	NotifyDriverNotFoundQueue = "notify_driver_not_found"

	NotifyDriverAssignedQueue = "notify_driver_assigned"
)

type DriverTripResponseData struct {
	Driver  *pb.Driver `json:"driver"`
	TripID  string     `json:"tripID"`
	RiderID string     `json:"riderID"`
}

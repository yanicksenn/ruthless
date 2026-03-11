package domain

import (
	"github.com/google/uuid"
	pb "github.com/yanicksenn/ruthless/api/v1"
)

func NewUser(name string) *pb.User {
	return &pb.User{
		Id:   uuid.New().String(),
		Name: name,
	}
}

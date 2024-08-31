package api

import (
	"common/rpc"
	"context"
	"github.com/gin-gonic/gin"
	"user/pb"
)

type UserHandler struct {
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (u *UserHandler) Register(ctx *gin.Context) {
	response, err := rpc.UserClient.Register(context.TODO(), &pb.RegisterParams{})
	if err != nil {

	}

	uid := response.Uid

}

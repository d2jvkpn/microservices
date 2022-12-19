package models

import (
	"context"
	// "fmt"

	. "authentication/proto"
)

// impls AuthServiceServer
type Server struct{}

func NewServer() *Server {
	return &Server{}
}

func (srv *Server) Create(ctx context.Context, in *CreateQ) (ans *CreateA, err error) {
	ans = &CreateA{
		Id:  "x001",
		Msg: &Msg{Code: 0, HttpCode: 200, Msg: "ok"},
	}

	return ans, nil
}

func (srv *Server) Verify(ctx context.Context, in *VerifyQ) (ans *VerifyA, err error) {
	ans = new(VerifyA)
	return ans, nil
}

func (srv *Server) GetOrUpdate(ctx context.Context, in *GetOrUpdateQ) (
	ans *GetOrUpdateA, err error) {

	ans = new(GetOrUpdateA)
	return ans, nil
}

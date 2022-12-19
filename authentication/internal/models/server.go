package models

import (
	"context"
	// "fmt"
	"net/http"

	. "authentication/proto"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	// "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// impls AuthServiceServer
type Server struct {
	/*...*/
}

func NewServer() *Server {
	return &Server{}
}

type User struct {
	Id     string `gorm:"column:id"`
	Bah    string `gorm:"column:bah"`
	Status string `gorm:"column:status"`
}

func (srv *Server) Create(ctx context.Context, in *CreateQ) (ans *CreateA, err error) {
	var (
		bts []byte
	)

	ans = &CreateA{
		Id:  "",
		Msg: &Msg{Code: 0, HttpCode: http.StatusOK, Msg: "ok"},
	}

	if in.Password == "" {
		ans.Msg = &Msg{Code: -1, HttpCode: http.StatusBadRequest, Msg: "invalid password"}
		return ans, status.Errorf(codes.InvalidArgument, ans.Msg.Msg)
	}
	// TODO: password validation

	if bts, err = bcrypt.GenerateFromPassword([]byte(in.Password), _BcryptCost); err != nil {
		ans.Msg = &Msg{
			Code:     1,
			HttpCode: http.StatusInternalServerError,
			Msg:      "failed to generate from password",
		}
		return ans, status.Errorf(codes.Internal, err.Error())
	}

	err = _DB.WithContext(ctx).
		Raw("insert into users (bah) values (?) returning id", string(bts)).
		Pluck("id", &ans.Id).Error
	if err != nil {
		ans.Msg = &Msg{
			Code:     2,
			HttpCode: http.StatusInternalServerError,
			Msg:      "failed to insert a record",
		}
		return ans, status.Errorf(codes.Internal, err.Error())
	}

	return ans, nil
}

func (srv *Server) Verify(ctx context.Context, in *VerifyQ) (ans *VerifyA, err error) {
	if in.Id == "" || in.Password == "" {
		ans.Msg = &Msg{Code: -1, HttpCode: http.StatusBadRequest, Msg: "invalid id or password"}
		return ans, status.Errorf(codes.InvalidArgument, ans.Msg.Msg)
	}

	var user User

	ans = &VerifyA{
		Status: "",
		Msg:    &Msg{Code: 0, HttpCode: http.StatusOK, Msg: "ok"},
	}

	err = _DB.WithContext(ctx).Table("users").
		Where("id = ?", in.Id).Limit(1).
		Select("bah, status").Find(&user).Error
	if err != nil {
		ans.Msg = &Msg{
			Code:     1,
			HttpCode: http.StatusInternalServerError,
			Msg:      "failed to retrieve",
		}
		return ans, status.Errorf(codes.Internal, err.Error())
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Bah), []byte(in.Password)); err != nil {
		ans.Msg = &Msg{
			Code:     2,
			HttpCode: http.StatusInternalServerError,
			Msg:      "compare password failed",
		}
		return ans, status.Errorf(codes.Internal, err.Error())
	}

	ans.Status = user.Status
	return ans, nil
}

func (srv *Server) GetOrUpdate(ctx context.Context, in *GetOrUpdateQ) (
	ans *GetOrUpdateA, err error) {

	ans = new(GetOrUpdateA)
	return ans, nil
}

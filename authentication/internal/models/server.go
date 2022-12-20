package models

import (
	"context"
	// "fmt"
	"net/http"

	"authentication/internal/settings"
	. "authentication/proto"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
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

func (srv *Server) Create_v0(ctx context.Context, in *CreateQ) (ans *CreateA, err error) {
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

	commonLabels := []attribute.KeyValue{
		attribute.Int("password.Length", len(in.Password)),
	}
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(commonLabels...)

	logger := settings.Logger.Named("trace")
	logger.Debug("Create",
		zap.String("traceId", span.SpanContext().TraceID().String()),
		zap.String("spanId", span.SpanContext().SpanID().String()),
		zap.Any("labels", commonLabels),
	)

	if bts, err = createGenerateFromPassword(ctx, in.Password, ans); err != nil {
		return nil, err
	}

	if err = createInsert(ctx, bts, ans); err != nil {
		return nil, err
	}

	return ans, nil
}

func createGenerateFromPassword(ctx context.Context, password string, ans *CreateA) (
	bts []byte, err error) {
	tracer := otel.Tracer(settings.App)
	_, span := tracer.Start(ctx, "bcrypt.GenerateFromPassword")
	defer span.End()

	logger := settings.Logger.Named("trace")
	logger.Debug("Create",
		zap.String("traceId", span.SpanContext().TraceID().String()),
		zap.String("spanId", span.SpanContext().SpanID().String()),
	)

	if bts, err = bcrypt.GenerateFromPassword([]byte(password), _BcryptCost); err != nil {
		ans.Msg = &Msg{
			Code:     1,
			HttpCode: http.StatusInternalServerError,
			Msg:      "failed to generate from password",
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return bts, nil
}

func createInsert(ctx context.Context, bts []byte, ans *CreateA) (err error) {
	tracer := otel.Tracer(settings.App)
	_, span := tracer.Start(ctx, "postgres.Insert")
	defer span.End()

	logger := settings.Logger.Named("trace")
	logger.Debug("Create",
		zap.String("traceId", span.SpanContext().TraceID().String()),
		zap.String("spanId", span.SpanContext().SpanID().String()),
	)

	err = _DB.WithContext(ctx).
		Raw("insert into users (bah) values (?) returning id", string(bts)).
		Pluck("id", &ans.Id).Error
	if err != nil {
		ans.Msg = &Msg{
			Code:     2,
			HttpCode: http.StatusInternalServerError,
			Msg:      "failed to insert a record",
		}

		return status.Errorf(codes.Internal, err.Error())
	}

	opts := []trace.EventOption{
		trace.WithAttributes(attribute.String("id", ans.Id)),
	}
	span.AddEvent("successfully finished Create", opts...)

	return nil
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

	ans = &GetOrUpdateA{
		Status: "",
		Msg:    &Msg{Code: 0, HttpCode: http.StatusOK, Msg: "ok"},
	}

	if in.Id == "" {
		ans.Msg = &Msg{Code: -1, HttpCode: http.StatusBadRequest, Msg: "invalid id"}
		return ans, status.Errorf(codes.InvalidArgument, ans.Msg.Msg)
	}

	tx := _DB.WithContext(ctx).Table("users").Where("id = ?", in.Id).Limit(1)

	switch {
	case in.Password != "" && in.Status != "":
		ans.Msg = &Msg{
			Code: -1, HttpCode: http.StatusBadRequest,
			Msg: "don't pass both password and status",
		}
		err = status.Errorf(codes.InvalidArgument, ans.Msg.Msg)
	case in.Password == "" && in.Status == "":
		if err = tx.Pluck("status", &ans.Status).Error; err == nil {
			break
		}

		if err.Error() == "record not found" {
			ans.Msg.Code, ans.Msg.Msg = -2, "failed to retrieve"
			ans.Msg.HttpCode = http.StatusNotFound
			err = status.Errorf(codes.NotFound, err.Error())
		} else {
			ans.Msg.Code, ans.Msg.Msg = 1, "record not found"
			ans.Msg.HttpCode = http.StatusInternalServerError
			err = status.Errorf(codes.Internal, err.Error())
		}
	case in.Password != "":
		var bts []byte
		if bts, err = bcrypt.GenerateFromPassword([]byte(in.Password), _BcryptCost); err != nil {
			ans.Msg = &Msg{
				Code:     2,
				HttpCode: http.StatusInternalServerError,
				Msg:      "failed to generate from password",
			}
			err = status.Errorf(codes.Internal, err.Error())
			break
		}

		if err = tx.Update("bah", string(bts)).Error; err != nil {
			ans.Msg = &Msg{
				Code:     3,
				HttpCode: http.StatusInternalServerError,
				Msg:      "failed to update",
			}
			err = status.Errorf(codes.Internal, err.Error())
		}
	default: // status != ""
		if err = tx.Update("status", in.Status).Error; err != nil {
			ans.Msg = &Msg{
				Code:     4,
				HttpCode: http.StatusInternalServerError,
				Msg:      "failed to update",
			}
			err = status.Errorf(codes.Internal, err.Error())
		}
	}

	return ans, err
}

package models

import (
	"context"
	"fmt"
	"net/http"

	"github.com/d2jvkpn/microservices/authentication/internal/settings"
	. "github.com/d2jvkpn/microservices/authentication/proto"

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
		Msg: &Msg{HttpCode: http.StatusOK, Msg: "ok"},
	}

	if in.Password == "" {
		ans.Msg = &Msg{HttpCode: http.StatusBadRequest, Msg: "invalid password"}
		return ans, status.Errorf(codes.InvalidArgument, ans.Msg.Msg)
	}
	// TODO: password validation

	if bts, err = bcrypt.GenerateFromPassword([]byte(in.Password), _BcryptCost); err != nil {
		ans.Msg = &Msg{
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
		Msg: &Msg{HttpCode: http.StatusOK, Msg: "ok"},
	}

	if in.Password == "" {
		ans.Msg = &Msg{HttpCode: http.StatusBadRequest, Msg: "invalid password"}
		return ans, status.Errorf(codes.InvalidArgument, ans.Msg.Msg)
	}
	// TODO: password validation

	commonLabels := []attribute.KeyValue{
		attribute.Int("password.Length", len(in.Password)),
	}
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(commonLabels...)

	if traceId := span.SpanContext().TraceID(); traceId.IsValid() {
		logger := settings.Logger.Named("trace")

		logger.Debug("models.Create",
			zap.String("traceId", traceId.String()),
			zap.String("spanId", span.SpanContext().SpanID().String()),
			zap.Any("labels", commonLabels),
		)
	}

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

	if traceId := span.SpanContext().TraceID(); traceId.IsValid() {
		logger := settings.Logger.Named("trace")
		logger.Debug("models.createGenerateFromPassword",
			zap.String("traceId", traceId.String()),
			zap.String("spanId", span.SpanContext().SpanID().String()),
		)
	}

	if bts, err = bcrypt.GenerateFromPassword([]byte(password), _BcryptCost); err != nil {
		ans.Msg = &Msg{
			HttpCode: http.StatusInternalServerError,
			Msg:      "failed to generate from password",
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return bts, nil
}

func createInsert(ctx context.Context, bts []byte, ans *CreateA) (err error) {
	opts := []trace.EventOption{
		trace.WithAttributes(attribute.String("id", ans.Id)),
	}

	tracer := otel.Tracer(settings.App)
	_, span := tracer.Start(ctx, "postgres.Insert")
	defer func() {
		if err == nil {
			span.AddEvent("successfully finished Create", opts...)
		} else {
			span.AddEvent(fmt.Sprintf("failed to Create: %v", err), opts...)
		}
		span.End()
	}()

	if traceId := span.SpanContext().TraceID(); traceId.IsValid() {
		logger := settings.Logger.Named("trace")
		logger.Debug("models.createInsert",
			zap.String("traceId", traceId.String()),
			zap.String("spanId", span.SpanContext().SpanID().String()),
		)
	}

	err = _DB.WithContext(ctx).
		Raw("insert into users (bah) values (?) returning id", string(bts)).
		Pluck("id", &ans.Id).Error
	if err != nil {
		ans.Msg = &Msg{
			HttpCode: http.StatusInternalServerError,
			Msg:      "failed to insert a record",
		}

		return status.Errorf(codes.Internal, err.Error())
	}

	return nil
}

func (srv *Server) Verify(ctx context.Context, in *VerifyQ) (ans *VerifyA, err error) {
	if in.Id == "" || in.Password == "" {
		ans.Msg = &Msg{HttpCode: http.StatusBadRequest, Msg: "invalid id or password"}
		return ans, status.Errorf(codes.InvalidArgument, ans.Msg.Msg)
	}

	var user User

	ans = &VerifyA{
		Status: "",
		Msg:    &Msg{HttpCode: http.StatusOK, Msg: "ok"},
	}

	err = _DB.WithContext(ctx).Table("users").
		Where("id = ?", in.Id).Limit(1).
		Select("bah, status").Find(&user).Error
	if err != nil {
		ans.Msg = &Msg{
			HttpCode: http.StatusInternalServerError,
			Msg:      "failed to retrieve",
		}
		return ans, status.Errorf(codes.Internal, err.Error())
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Bah), []byte(in.Password)); err != nil {
		ans.Msg = &Msg{
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

	var (
		bts   []byte
		event string
	)

	ans = &GetOrUpdateA{
		Status: "",
		Msg:    &Msg{HttpCode: http.StatusOK, Msg: "ok"},
	}

	if in.Id == "" {
		ans.Msg = &Msg{HttpCode: http.StatusBadRequest, Msg: "invalid id"}
		return ans, status.Errorf(codes.InvalidArgument, ans.Msg.Msg)
	}

	if in.Password != "" && in.Status != "" {
		ans.Msg = &Msg{
			HttpCode: http.StatusBadRequest,
			Msg:      "don't pass both password and status",
		}
		return ans, status.Errorf(codes.InvalidArgument, ans.Msg.Msg)
	}

	opts := []trace.EventOption{
		trace.WithAttributes(attribute.String("id", in.Id)),
	}
	tracer := otel.Tracer(settings.App)
	_, span := tracer.Start(ctx, "models.GetOrUpdate")
	defer func() {
		if err == nil {
			span.AddEvent(fmt.Sprintf("successfully finished to %s", event), opts...)
		} else {
			span.AddEvent(fmt.Sprintf("failed to %s: %v", event, err), opts...)
		}
		span.End()
	}()

	tx := _DB.WithContext(ctx).Table("users").Where("id = ?", in.Id).Limit(1)

	switch {
	case in.Password == "" && in.Status == "":
		event = "get"
		if err = tx.Pluck("status", &ans.Status).Error; err == nil {
			break
		}

		if err.Error() == "record not found" {
			ans.Msg.Msg = "failed to retrieve"
			ans.Msg.HttpCode = http.StatusNotFound
			err = status.Errorf(codes.NotFound, err.Error())
		} else {
			ans.Msg.Msg = "record not found"
			ans.Msg.HttpCode = http.StatusInternalServerError
			err = status.Errorf(codes.Internal, err.Error())
		}
	case in.Password != "":
		event = "update password"
		if bts, err = bcrypt.GenerateFromPassword([]byte(in.Password), _BcryptCost); err != nil {
			ans.Msg = &Msg{
				HttpCode: http.StatusInternalServerError,
				Msg:      "failed to generate from password",
			}
			err = status.Errorf(codes.Internal, err.Error())
			break
		}

		if err = tx.Update("bah", string(bts)).Error; err != nil {
			ans.Msg = &Msg{
				HttpCode: http.StatusInternalServerError,
				Msg:      "failed to update",
			}
			err = status.Errorf(codes.Internal, err.Error())
		}
	default: // status != ""
		event = "update status"
		if err = tx.Update("status", in.Status).Error; err != nil {
			ans.Msg = &Msg{
				HttpCode: http.StatusInternalServerError,
				Msg:      "failed to update",
			}
			err = status.Errorf(codes.Internal, err.Error())
		}
	}

	return ans, err
}

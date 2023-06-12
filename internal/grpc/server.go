package grpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	"github.com/v8tix/mallbots-payments-proto/pb"
	"github.com/v8tix/mallbots-payments/internal/application"
)

type server struct {
	app application.App
	pb.UnimplementedPaymentsServiceServer
}

var _ pb.PaymentsServiceServer = (*server)(nil)

func RegisterServer(_ context.Context, app application.App, registrar grpc.ServiceRegistrar) error {
	pb.RegisterPaymentsServiceServer(registrar, server{app: app})
	return nil
}

func (s server) AuthorizePayment(ctx context.Context, request *pb.AuthorizePaymentRequest,
) (*pb.AuthorizePaymentResponse, error) {
	id := uuid.New().String()
	err := s.app.AuthorizePayment(ctx, application.AuthorizePayment{
		ID:         id,
		CustomerID: request.GetCustomerId(),
		Amount:     request.GetAmount(),
	})
	return &pb.AuthorizePaymentResponse{Id: id}, err
}

func (s server) ConfirmPayment(ctx context.Context, request *pb.ConfirmPaymentRequest,
) (*pb.ConfirmPaymentResponse, error) {
	err := s.app.ConfirmPayment(ctx, application.ConfirmPayment{
		ID: request.GetId(),
	})
	return &pb.ConfirmPaymentResponse{}, err
}

func (s server) CreateInvoice(ctx context.Context, request *pb.CreateInvoiceRequest,
) (*pb.CreateInvoiceResponse, error) {
	id := uuid.New().String()
	err := s.app.CreateInvoice(ctx, application.CreateInvoice{
		ID:      id,
		OrderID: request.GetOrderId(),
		Amount:  request.GetAmount(),
	})
	return &pb.CreateInvoiceResponse{
		Id: id,
	}, err
}

func (s server) AdjustInvoice(ctx context.Context, request *pb.AdjustInvoiceRequest,
) (*pb.AdjustInvoiceResponse, error) {
	err := s.app.AdjustInvoice(ctx, application.AdjustInvoice{
		ID:     request.GetId(),
		Amount: request.GetAmount(),
	})
	return &pb.AdjustInvoiceResponse{}, err
}

func (s server) PayInvoice(ctx context.Context, request *pb.PayInvoiceRequest) (*pb.PayInvoiceResponse,
	error,
) {
	err := s.app.PayInvoice(ctx, application.PayInvoice{
		ID: request.GetId(),
	})
	return &pb.PayInvoiceResponse{}, err
}

func (s server) CancelInvoice(ctx context.Context, request *pb.CancelInvoiceRequest,
) (*pb.CancelInvoiceResponse, error) {
	err := s.app.CancelInvoice(ctx, application.CancelInvoice{
		ID: request.GetId(),
	})
	return &pb.CancelInvoiceResponse{}, err
}

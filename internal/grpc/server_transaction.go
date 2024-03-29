package grpc

import (
	"context"
	"database/sql"

	"google.golang.org/grpc"

	"github.com/v8tix/eda/di"
	"github.com/v8tix/mallbots-payments-proto/pb"
	"github.com/v8tix/mallbots-payments/internal/application"
)

type serverTx struct {
	c di.Container
	pb.UnimplementedPaymentsServiceServer
}

var _ pb.PaymentsServiceServer = (*serverTx)(nil)

func RegisterServerTx(container di.Container, registrar grpc.ServiceRegistrar) error {
	pb.RegisterPaymentsServiceServer(registrar, serverTx{
		c: container,
	})
	return nil
}

func (s serverTx) AuthorizePayment(ctx context.Context, request *pb.AuthorizePaymentRequest) (resp *pb.AuthorizePaymentResponse, err error) {
	ctx = s.c.Scoped(ctx)
	defer func(tx *sql.Tx) {
		err = s.closeTx(tx, err)
	}(di.Get(ctx, "tx").(*sql.Tx))

	next := server{app: di.Get(ctx, "app").(application.App)}

	return next.AuthorizePayment(ctx, request)
}

func (s serverTx) ConfirmPayment(ctx context.Context, request *pb.ConfirmPaymentRequest) (resp *pb.ConfirmPaymentResponse, err error) {
	ctx = s.c.Scoped(ctx)
	defer func(tx *sql.Tx) {
		err = s.closeTx(tx, err)
	}(di.Get(ctx, "tx").(*sql.Tx))

	next := server{app: di.Get(ctx, "app").(application.App)}

	return next.ConfirmPayment(ctx, request)
}

func (s serverTx) CreateInvoice(ctx context.Context, request *pb.CreateInvoiceRequest) (resp *pb.CreateInvoiceResponse, err error) {
	ctx = s.c.Scoped(ctx)
	defer func(tx *sql.Tx) {
		err = s.closeTx(tx, err)
	}(di.Get(ctx, "tx").(*sql.Tx))

	next := server{app: di.Get(ctx, "app").(application.App)}

	return next.CreateInvoice(ctx, request)
}

func (s serverTx) AdjustInvoice(ctx context.Context, request *pb.AdjustInvoiceRequest) (resp *pb.AdjustInvoiceResponse, err error) {
	ctx = s.c.Scoped(ctx)
	defer func(tx *sql.Tx) {
		err = s.closeTx(tx, err)
	}(di.Get(ctx, "tx").(*sql.Tx))

	next := server{app: di.Get(ctx, "app").(application.App)}

	return next.AdjustInvoice(ctx, request)
}

func (s serverTx) PayInvoice(ctx context.Context, request *pb.PayInvoiceRequest) (resp *pb.PayInvoiceResponse, err error) {
	ctx = s.c.Scoped(ctx)
	defer func(tx *sql.Tx) {
		err = s.closeTx(tx, err)
	}(di.Get(ctx, "tx").(*sql.Tx))

	next := server{app: di.Get(ctx, "app").(application.App)}

	return next.PayInvoice(ctx, request)
}

func (s serverTx) CancelInvoice(ctx context.Context, request *pb.CancelInvoiceRequest) (resp *pb.CancelInvoiceResponse, err error) {
	ctx = s.c.Scoped(ctx)
	defer func(tx *sql.Tx) {
		err = s.closeTx(tx, err)
	}(di.Get(ctx, "tx").(*sql.Tx))

	next := server{app: di.Get(ctx, "app").(application.App)}

	return next.CancelInvoice(ctx, request)
}

func (s serverTx) closeTx(tx *sql.Tx, err error) error {
	if p := recover(); p != nil {
		_ = tx.Rollback()
		panic(p)
	} else if err != nil {
		_ = tx.Rollback()
		return err
	} else {
		return tx.Commit()
	}
}

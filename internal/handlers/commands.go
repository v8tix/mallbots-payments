package handlers

import (
	"context"

	"github.com/v8tix/eda/am"
	"github.com/v8tix/eda/ddd"
	"github.com/v8tix/mallbots-payments-proto/pb"
	"github.com/v8tix/mallbots-payments/internal/application"
)

type commandHandlers struct {
	app application.App
}

func NewCommandHandlers(app application.App) ddd.CommandHandler[ddd.Command] {
	return commandHandlers{
		app: app,
	}
}

func RegisterCommandHandlers(subscriber am.RawMessageSubscriber, handlers am.RawMessageHandler) error {
	return subscriber.Subscribe(pb.CommandChannel, handlers, am.MessageFilter{
		pb.ConfirmPaymentCommand,
	}, am.GroupName("payment-commands"))
}

func (h commandHandlers) HandleCommand(ctx context.Context, cmd ddd.Command) (ddd.Reply, error) {
	switch cmd.CommandName() {
	case pb.ConfirmPaymentCommand:
		return h.doConfirmPayment(ctx, cmd)
	}

	return nil, nil
}

func (h commandHandlers) doConfirmPayment(ctx context.Context, cmd ddd.Command) (ddd.Reply, error) {
	payload := cmd.Payload().(*pb.ConfirmPayment)

	return nil, h.app.ConfirmPayment(ctx, application.ConfirmPayment{ID: payload.GetId()})
}

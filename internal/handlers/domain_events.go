package handlers

import (
	"context"

	"github.com/v8tix/eda/am"
	"github.com/v8tix/eda/ddd"
	"github.com/v8tix/mallbots-payments-proto/pb"
	"github.com/v8tix/mallbots-payments/internal/models"
)

type domainHandlers[T ddd.Event] struct {
	publisher am.MessagePublisher[ddd.Event]
}

var _ ddd.EventHandler[ddd.Event] = (*domainHandlers[ddd.Event])(nil)

func NewDomainEventHandlers(publisher am.MessagePublisher[ddd.Event]) ddd.EventHandler[ddd.Event] {
	return &domainHandlers[ddd.Event]{
		publisher: publisher,
	}
}

func RegisterDomainEventHandlers(subscriber ddd.EventSubscriber[ddd.Event], handlers ddd.EventHandler[ddd.Event]) {
	subscriber.Subscribe(handlers,
		models.InvoicePaidEvent,
	)
}

func (h domainHandlers[T]) HandleEvent(ctx context.Context, event T) error {
	switch event.EventName() {
	case models.InvoicePaidEvent:
		return h.onInvoicePaid(ctx, event)
	}
	return nil
}

func (h domainHandlers[T]) onInvoicePaid(ctx context.Context, event ddd.Event) error {
	payload := event.Payload().(*models.InvoicePaid)
	return h.publisher.Publish(ctx, pb.InvoiceAggregateChannel,
		ddd.NewEvent(pb.InvoicePaidEvent, &pb.InvoicePaid{
			Id:      payload.ID,
			OrderId: payload.OrderID,
		}),
	)
}

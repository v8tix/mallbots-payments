package payments

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog"

	"github.com/v8tix/eda/am"
	"github.com/v8tix/eda/ddd"
	"github.com/v8tix/eda/di"
	"github.com/v8tix/eda/jetstream"
	pg "github.com/v8tix/eda/postgres"
	"github.com/v8tix/eda/registry"
	"github.com/v8tix/eda/tm"
	orderingpb "github.com/v8tix/mallbots-ordering-proto/pb"
	"github.com/v8tix/mallbots-payments-proto/pb"
	"github.com/v8tix/mallbots-payments-proto/rest"
	"github.com/v8tix/mallbots-payments/internal/application"
	"github.com/v8tix/mallbots-payments/internal/grpc"
	"github.com/v8tix/mallbots-payments/internal/handlers"
	"github.com/v8tix/mallbots-payments/internal/logging"
	"github.com/v8tix/mallbots-payments/internal/ms"
	"github.com/v8tix/mallbots-payments/internal/postgres"
)

type Module struct{}

func (m Module) Startup(ctx context.Context, mono ms.Microservice) (err error) {
	container := di.New()
	// setup Driven adapters
	container.AddSingleton("registry", func(c di.Container) (any, error) {
		reg := registry.New()
		if err := orderingpb.Registrations(reg); err != nil {
			return nil, err
		}
		if err := pb.Registrations(reg); err != nil {
			return nil, err
		}
		return reg, nil
	})
	container.AddSingleton("logger", func(c di.Container) (any, error) {
		return mono.Logger(), nil
	})
	container.AddSingleton("stream", func(c di.Container) (any, error) {
		return jetstream.NewStream(mono.Config().Nats.Stream, mono.JS(), c.Get("logger").(zerolog.Logger)), nil
	})
	container.AddSingleton("domainDispatcher", func(c di.Container) (any, error) {
		return ddd.NewEventDispatcher[ddd.Event](), nil
	})
	container.AddSingleton("db", func(c di.Container) (any, error) {
		return mono.DB(), nil
	})
	container.AddSingleton("outboxProcessor", func(c di.Container) (any, error) {
		return tm.NewOutboxProcessor(
			c.Get("stream").(am.RawMessageStream),
			pg.NewOutboxStore("payments.outbox", c.Get("db").(*sql.DB)),
		), nil
	})
	container.AddScoped("tx", func(c di.Container) (any, error) {
		db := c.Get("db").(*sql.DB)
		return db.Begin()
	})
	container.AddScoped("txStream", func(c di.Container) (any, error) {
		tx := c.Get("tx").(*sql.Tx)
		outboxStore := pg.NewOutboxStore("payments.outbox", tx)
		return am.RawMessageStreamWithMiddleware(
			c.Get("stream").(am.RawMessageStream),
			tm.NewOutboxStreamMiddleware(outboxStore),
		), nil
	})
	container.AddScoped("eventStream", func(c di.Container) (any, error) {
		return am.NewEventStream(c.Get("registry").(registry.Registry), c.Get("txStream").(am.RawMessageStream)), nil
	})
	container.AddScoped("replyStream", func(c di.Container) (any, error) {
		return am.NewReplyStream(c.Get("registry").(registry.Registry), c.Get("txStream").(am.RawMessageStream)), nil
	})
	container.AddScoped("inboxMiddleware", func(c di.Container) (any, error) {
		tx := c.Get("tx").(*sql.Tx)
		inboxStore := pg.NewInboxStore("payments.inbox", tx)
		return tm.NewInboxHandlerMiddleware(inboxStore), nil
	})
	container.AddScoped("invoices", func(c di.Container) (any, error) {
		return postgres.NewInvoiceRepository("payments.invoices", c.Get("tx").(*sql.Tx)), nil
	})
	container.AddScoped("payments", func(c di.Container) (any, error) {
		return postgres.NewPaymentRepository("payments.payments", c.Get("tx").(*sql.Tx)), nil
	})

	// setup application
	container.AddScoped("app", func(c di.Container) (any, error) {
		return logging.LogApplicationAccess(
			application.New(
				c.Get("invoices").(application.InvoiceRepository),
				c.Get("payments").(application.PaymentRepository),
				c.Get("domainDispatcher").(*ddd.EventDispatcher[ddd.Event]),
			),
			c.Get("logger").(zerolog.Logger),
		), nil
	})
	container.AddScoped("domainEventHandlers", func(c di.Container) (any, error) {
		return logging.LogEventHandlerAccess[ddd.Event](
			handlers.NewDomainEventHandlers(c.Get("eventStream").(am.EventStream)),
			"DomainEvents", c.Get("logger").(zerolog.Logger),
		), nil
	})
	container.AddScoped("integrationEventHandlers", func(c di.Container) (any, error) {
		return logging.LogEventHandlerAccess[ddd.Event](
			handlers.NewIntegrationHandlers(
				c.Get("app").(application.App),
			),
			"IntegrationEvents", c.Get("logger").(zerolog.Logger),
		), nil
	})
	container.AddScoped("commandHandlers", func(c di.Container) (any, error) {
		return logging.LogCommandHandlerAccess[ddd.Command](
			handlers.NewCommandHandlers(c.Get("app").(application.App)),
			"Commands", c.Get("logger").(zerolog.Logger),
		), nil
	})

	// setup Driver adapters
	if err = grpc.RegisterServerTx(container, mono.RPC()); err != nil {
		return err
	}
	if err = rest.RegisterGateway(ctx, mono.Mux(), mono.Config().RPC.Address()); err != nil {
		return err
	}
	if err = rest.RegisterSwagger(mono.Mux()); err != nil {
		return err
	}
	if err = handlers.RegisterIntegrationEventHandlersTx(container); err != nil {
		return err
	}
	handlers.RegisterDomainEventHandlersTx(container)
	if err = handlers.RegisterCommandHandlersTx(container); err != nil {
		return err
	}
	startOutboxProcessor(ctx, container)

	return
}

func startOutboxProcessor(ctx context.Context, container di.Container) {
	outboxProcessor := container.Get("outboxProcessor").(tm.OutboxProcessor)
	logger := container.Get("logger").(zerolog.Logger)

	go func() {
		err := outboxProcessor.Start(ctx)
		if err != nil {
			logger.Error().Err(err).Msg("payments outbox processor encountered an error")
		}
	}()
}

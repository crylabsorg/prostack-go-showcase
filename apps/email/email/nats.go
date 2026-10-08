package email

import (
	"context"
	"encoding/json/v2"
	"errors"
	"time"

	"github.com/crylabsorg/prostack-go-showcase/pkg/events"
	"github.com/crylabsorg/prostack-go/logger"
	"github.com/crylabsorg/prostack-go/ps_nats"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

type natsController struct {
	natsConn *nats.Conn
	service  *service
}

func Run(ctx context.Context, drainingCtx context.Context, natsConn *nats.Conn, sender Sender) {
	controller := &natsController{natsConn: natsConn, service: &service{sender: sender}}
	go controller.run(ctx, drainingCtx)
}

func (c *natsController) run(ctx context.Context, drainingCtx context.Context) {
	for {
		consumer, err := c.createConsumer(ctx)
		if err == nil {
			err = ps_nats.RunConsumer(ctx, drainingCtx, consumer, 50, c.handleMessage)
		}
		if err != nil {
			logger.GetLogger(ctx).Error("email consumer", zap.Error(err))
		}
		select {
		case <-ctx.Done():
			return
		case <-drainingCtx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func (c *natsController) createConsumer(ctx context.Context) (jetstream.Consumer, error) {
	js, err := jetstream.New(c.natsConn)
	if err != nil {
		return nil, err
	}
	consumerCfg := jetstream.ConsumerConfig{
		Durable:       "email-sender",
		FilterSubject: events.UserCreatedSubject,
	}
	return js.CreateOrUpdateConsumer(ctx, events.UserStream, consumerCfg)
}

func (c *natsController) handleMessage(ctx context.Context, msg jetstream.Msg) error {
	var event events.UserCreatedEvent
	err := json.Unmarshal(msg.Data(), &event)
	if err != nil {
		if termErr := msg.Term(); termErr != nil {
			err = errors.Join(err, termErr)
		}
		return err
	}
	err = c.service.sendWelcome(ctx, SendWelcomeInput{ID: event.ID, Name: event.Name, Email: event.Email})
	if err == nil {
		err = msg.DoubleAck(ctx)
	} else {
		if nakErr := msg.Nak(); nakErr != nil {
			err = errors.Join(err, nakErr)
		}
	}
	if err != nil {
		logger.GetLogger(ctx).Error("send welcome email", zap.Error(err))
	}
	return err
}

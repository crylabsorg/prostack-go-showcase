package email

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/crylabsorg/prostack-go-showcase/pkg/events"
	"github.com/crylabsorg/prostack-go/logger"
	"github.com/crylabsorg/prostack-go/ps_nats"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

type natsController struct {
	client  *nats.Conn
	service *service
}

func Run(ctx context.Context, client *nats.Conn, sender Sender) {
	(&natsController{client: client, service: &service{sender: sender}}).run(ctx)
}

func (c *natsController) run(ctx context.Context) {
	for ctx.Err() == nil {
		if err := c.consume(ctx); err != nil && ctx.Err() == nil {
			logger.GetLogger(ctx).Error("email consumer", zap.Error(err))
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func (c *natsController) consume(ctx context.Context) error {
	js, err := jetstream.New(c.client)
	if err != nil {
		return err
	}
	consumer, err := js.CreateOrUpdateConsumer(ctx, events.UserStream, jetstream.ConsumerConfig{
		Durable: "email-sender", FilterSubject: events.UserCreatedSubject,
		AckPolicy: jetstream.AckExplicitPolicy, AckWait: 30 * time.Second,
	})
	if err != nil {
		return err
	}
	slots := make(chan struct{}, 50)
	for range cap(slots) {
		slots <- struct{}{}
	}
	var workers sync.WaitGroup
	defer workers.Wait()
	for ctx.Err() == nil {
		if len(slots) == 0 {
			select {
			case <-ctx.Done():
				return nil
			case <-slots:
				slots <- struct{}{}
			}
		}
		batch, err := consumer.Fetch(len(slots), jetstream.FetchMaxWait(time.Second))
		if err != nil {
			return err
		}
		for msg := range batch.Messages() {
			<-slots
			workers.Go(func() {
				defer func() { slots <- struct{}{} }()
				c.handleMessage(ctx, msg)
			})
		}
		if err := batch.Error(); err != nil {
			return err
		}
	}
	return nil
}

func (c *natsController) handleMessage(ctx context.Context, msg jetstream.Msg) {
	msgCtx, span := ps_nats.TraceContext(ctx, msg)
	defer span.End()
	var event events.UserCreated
	err := json.Unmarshal(msg.Data(), &event)
	if err != nil {
		if termErr := msg.Term(); termErr != nil {
			err = errors.Join(err, termErr)
		}
	} else {
		err = c.service.sendWelcome(msgCtx, Welcome{ID: event.ID, Name: event.Name, Email: event.Email})
		if err == nil {
			err = msg.DoubleAck(msgCtx)
		} else {
			if nakErr := msg.NakWithDelay(30 * time.Second); nakErr != nil {
				err = errors.Join(err, nakErr)
			}
		}
	}
	if err != nil {
		span.RecordError(err)
		logger.GetLogger(msgCtx).Error("send welcome email", zap.Error(err))
	}
}

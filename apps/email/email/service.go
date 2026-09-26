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
	"go.uber.org/zap"
)

type Sender interface {
	Send(context.Context, Welcome) error
}

type service struct {
	client *nats.Conn
	sender Sender
}

func newService(client *nats.Conn, sender Sender) *service {
	return &service{client: client, sender: sender}
}

func Run(ctx context.Context, client *nats.Conn, sender Sender) {
	newService(client, sender).run(ctx)
}

func (s *service) run(ctx context.Context) {
	for ctx.Err() == nil {
		if err := s.consume(ctx); err != nil && ctx.Err() == nil {
			logger.GetLogger(ctx).Error("email consumer", zap.Error(err))
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func (s *service) consume(ctx context.Context) error {
	js, err := s.client.JetStream()
	if err != nil {
		return err
	}
	sub, err := js.PullSubscribe(events.UserCreatedSubject, "email-sender", nats.BindStream(events.UserStream), nats.ManualAck(), nats.AckWait(30*time.Second))
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
		messages, err := sub.Fetch(len(slots), nats.MaxWait(time.Second))
		for _, msg := range messages {
			<-slots
			workers.Go(func() {
				defer func() { slots <- struct{}{} }()
				s.handleMessage(ctx, msg)
			})
		}
		if err != nil && !errors.Is(err, nats.ErrTimeout) {
			return err
		}
	}
	return nil
}

func (s *service) handleMessage(ctx context.Context, msg *nats.Msg) {
	msgCtx, span := ps_nats.TraceContext(ctx, msg)
	defer span.End()
	var event events.UserCreated
	err := json.Unmarshal(msg.Data, &event)
	if err != nil {
		if termErr := msg.Term(); termErr != nil {
			err = errors.Join(err, termErr)
		}
	} else {
		sendCtx, cancel := context.WithTimeout(msgCtx, 20*time.Second)
		err = s.sender.Send(sendCtx, Welcome{ID: event.ID, Name: event.Name, Email: event.Email})
		cancel()
		if err == nil {
			err = msg.AckSync()
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

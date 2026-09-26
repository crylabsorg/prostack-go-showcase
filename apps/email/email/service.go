package email

import (
	"context"
	"encoding/json"
	"fmt"
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

func Run(ctx context.Context, url string, sender Sender) {
	for ctx.Err() == nil {
		if err := consume(ctx, url, sender); err != nil && ctx.Err() == nil {
			logger.GetLogger(ctx).Error("email consumer", zap.Error(err))
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func consume(ctx context.Context, url string, sender Sender) error {
	client, err := ps_nats.Connect(url)
	if err != nil {
		return err
	}
	defer client.Close()
	js, err := client.JetStream()
	if err != nil {
		return err
	}
	sub, err := js.PullSubscribe(events.UserCreatedSubject, "email-sender", nats.BindStream(events.UserStream), nats.ManualAck(), nats.AckWait(30*time.Second))
	if err != nil {
		return err
	}
	for ctx.Err() == nil {
		messages, err := sub.Fetch(1, nats.MaxWait(time.Second))
		if err == nats.ErrTimeout {
			continue
		}
		if err != nil {
			return err
		}
		for _, msg := range messages {
			msgCtx, span := ps_nats.TraceContext(ctx, msg)
			var event events.UserCreated
			err = json.Unmarshal(msg.Data, &event)
			if err == nil && (event.ID == "" || event.Name == "" || event.Email == "") {
				err = fmt.Errorf("missing user.created fields")
			}
			if err != nil {
				_ = msg.Term()
			} else {
				sendCtx, cancel := context.WithTimeout(msgCtx, 20*time.Second)
				err = sender.Send(sendCtx, Welcome{ID: event.ID, Name: event.Name, Email: event.Email})
				cancel()
				if err == nil {
					err = msg.AckSync()
				} else {
					_ = msg.NakWithDelay(30 * time.Second)
				}
			}
			if err != nil {
				span.RecordError(err)
				logger.GetLogger(msgCtx).Error("send welcome email", zap.Error(err))
			}
			span.End()
		}
	}
	return nil
}

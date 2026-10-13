package main

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Foo struct{}

func (*Foo) A() {}
func (*Foo) B() {}
func (*Foo) C() {}

type AB interface {
	A()
	B()
}

type BC interface {
	B()
	C()
}

func main() {
	var f AB = &Foo{}
	y := f.(BC)
	// y.A()
	_ = y
	g, ctx := errgroup.WithContext(context.Background())
}

// func main() {
// 	nums := []int{1, 2, 3}

// 	addNum(nums[0:2])
// 	fmt.Println(nums)

// 	addNums(nums[0:2])
// 	fmt.Println(nums)
// }

// func addNum(nums []int) {
// 	nums = append(nums, 4)
// }

// func addNums(nums []int) {
// 	nums = append(nums, 5, 6)
// }

// import (
// 	"context"

// 	"github.com/crylabsorg/prostack-go-showcase/pkg/events"
// 	"github.com/crylabsorg/prostack-go-showcase/user/config"
// 	"github.com/crylabsorg/prostack-go-showcase/user/user"
// 	"github.com/crylabsorg/prostack-go/ps_bootstrap"
// 	"github.com/crylabsorg/prostack-go/ps_event"
// 	"github.com/crylabsorg/prostack-go/ps_kafka"
// 	"github.com/crylabsorg/prostack-go/ps_nats"
// 	"github.com/crylabsorg/prostack-go/ps_pg"
// 	"github.com/nats-io/nats.go/jetstream"
// )

// func main() {
// 	ps_bootstrap.Run(config.New, setup)
// }

// func setup(ctx context.Context, drainingCtx context.Context, cfg *config.Config, app *ps_bootstrap.App) error {
// 	databasePool, disconnectDatabase := ps_pg.New(ctx, cfg.DatabaseURL)
// 	app.RegisterCleanup(func() error {
// 		disconnectDatabase()
// 		return nil
// 	})
// 	app.RegisterHealthProbe(ctx, "postgres", databasePool.Ping)

// 	userRepository := user.NewRepository(databasePool)
// 	userMetrics := user.NewMetrics(app.GetMetricsRegistry())
// 	userService := user.NewService(userRepository, userMetrics)
// 	natsConn, err := ps_nats.Connect(cfg.NATSURL)
// 	if err != nil {
// 		return err
// 	}
// 	app.RegisterCleanup(func() error {
// 		natsConn.Close()
// 		return nil
// 	})
// 	jetStream, err := jetstream.New(natsConn)
// 	if err != nil {
// 		return err
// 	}
// 	if err = ps_nats.EnsureStream(ctx, jetStream, events.UserStream, "user.*"); err != nil {
// 		return err
// 	}

// 	kafkaPublisher, err := ps_kafka.NewPublisher(cfg.KafkaBrokers...)
// 	if err != nil {
// 		return err
// 	}
// 	app.RegisterCleanup(func() error {
// 		kafkaPublisher.Close()
// 		return nil
// 	})
// 	app.RegisterHealthProbe(ctx, "kafka", kafkaPublisher.Ping)
// 	publisherDone := ps_event.RunOutboxPublisher(ctx, drainingCtx, ps_event.Fanout{ps_nats.NewPublisher(jetStream), kafkaPublisher}, user.NewOutbox(userService))

// 	app.RegisterCleanup(func() error {
// 		<-publisherDone
// 		return nil
// 	})

// 	user.RegisterGRPC(app.GetGRPCServer(), userService)
// 	return nil
// }

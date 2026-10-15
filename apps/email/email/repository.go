package email

import (
	"context"
	"errors"
	"time"

	"github.com/crylabsorg/prostack-go/logger"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
	"go.uber.org/zap"
)

type Repository struct{ deliveries *mongo.Collection }

func NewRepository(ctx context.Context, db *mongo.Database) (*Repository, error) {
	// Make collection creation visible before any consumer starts a transaction.
	initializationDB := db.Client().Database(db.Name(), options.Database().SetWriteConcern(writeconcern.Majority()))
	err := initializationDB.CreateCollection(ctx, "welcome_deliveries")
	if err != nil {
		const namespaceExistsCode = 48
		commandErr, ok := errors.AsType[mongo.CommandError](err)
		if !ok || commandErr.Code != namespaceExistsCode {
			return nil, err
		}
	}
	return &Repository{deliveries: db.Collection("welcome_deliveries")}, nil
}

// withinTransaction deliberately does not retry the callback: it may send an email.
func (r *Repository) withinTransaction(ctx context.Context, fn func(context.Context) error) error {
	session, err := r.deliveries.Database().Client().StartSession()
	if err != nil {
		return err
	}
	defer func() {
		const cleanupTimeout = 5 * time.Second
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cleanupTimeout)
		defer cancel()
		if session.TransactionRunning() {
			if abortErr := session.AbortTransaction(cleanupCtx); abortErr != nil {
				logger.GetLogger(ctx).Error("abort email transaction", zap.Error(abortErr))
			}
		}
		session.EndSession(cleanupCtx)
	}()
	if err = session.StartTransaction(); err != nil {
		return err
	}
	return mongo.WithSession(ctx, session, func(transactionCtx context.Context) error {
		if callbackErr := fn(transactionCtx); callbackErr != nil {
			return callbackErr
		}
		return session.CommitTransaction(transactionCtx)
	})
}

// reserveWelcome must run in withinTransaction. The unique _id arbitrates concurrent sends.
func (r *Repository) reserveWelcome(ctx context.Context, userID string) (bool, error) {
	filter := bson.M{"_id": userID}
	err := r.deliveries.FindOne(ctx, filter).Err()
	if err == nil {
		return false, nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return false, err
	}
	_, err = r.deliveries.InsertOne(ctx, modelWelcomeDelivery{UserID: userID})
	return err == nil, err
}

func (r *Repository) completeWelcome(ctx context.Context, userID string) error {
	_, err := r.deliveries.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{"sent_at": time.Now()}})
	return err
}

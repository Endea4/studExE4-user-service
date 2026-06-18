package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/Endea4/studExE4-shared/events"
	"github.com/Endea4/studExE4-user-service/internal/models"
	"github.com/Endea4/studExE4-user-service/internal/repository"
	"github.com/IBM/sarama"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TripEventConsumer struct {
	debtRepo *repository.DebtRepository
	userRepo *repository.UserRepository
}

func NewTripEventConsumer(debtRepo *repository.DebtRepository, userRepo *repository.UserRepository) *TripEventConsumer {
	return &TripEventConsumer{debtRepo: debtRepo, userRepo: userRepo}
}

func (c *TripEventConsumer) Start(ctx context.Context, brokers []string) {
	consumer, err := events.NewConsumerGroup(brokers, "studex-user-service")
	if err != nil {
		log.Fatalf("Kafka consumer init failed: %v", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				consumer.Close()
				return
			default:
			}
			handler := &events.ConsumerHandler{
				Handle: func(msg *sarama.ConsumerMessage) error {
					var evt events.Event
					if err := json.Unmarshal(msg.Value, &evt); err != nil {
						return err
					}
					switch evt.Type {
					case events.EventTripCompleted:
						c.handleTripCompleted(evt)
					case events.EventTripCancelled:
						c.handleTripCancelled(evt)
					}
					return nil
				},
			}
			if err := consumer.Consume(ctx, []string{events.TopicTripEvents}, handler); err != nil {
				log.Printf("Kafka consumer error: %v", err)
				time.Sleep(5 * time.Second)
			}
		}
	}()

	log.Printf("Kafka consumer started: topic=%s group=studex-user-service", events.TopicTripEvents)
}

func (c *TripEventConsumer) StartRedis(ctx context.Context, rdb *redis.Client) {
	ch, err := events.Subscribe(ctx, rdb, events.RedisChannelTripEvents)
	if err != nil {
		log.Printf("Redis trip subscriber failed (will retry): %v", err)
		time.Sleep(5 * time.Second)
		ch, err = events.Subscribe(ctx, rdb, events.RedisChannelTripEvents)
		if err != nil {
			log.Printf("Redis trip subscriber failed again: %v", err)
			return
		}
	}
	log.Printf("Redis trip consumer started: channel=%s", events.RedisChannelTripEvents)

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch.Ch:
			if !ok {
				return
			}
			var evt events.Event
			if err := json.Unmarshal([]byte(msg.Payload), &evt); err != nil {
				log.Printf("redis trip consumer: unmarshal error: %v", err)
				continue
			}
			switch evt.Type {
			case events.EventTripCompleted:
				c.handleTripCompleted(evt)
			case events.EventTripCancelled:
				c.handleTripCancelled(evt)
			}
		}
	}
}

func (c *TripEventConsumer) handleTripCompleted(evt events.Event) {
	dataBytes, _ := json.Marshal(evt.Data)
	var data events.TripCompletedData
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		log.Printf("failed to unmarshal trip.completed: %v", err)
		return
	}

	commissionRate := 0.10
	driverPayout := data.FinalPrice * (1 - commissionRate)

	driverPhone := data.DriverRefID
	if user, err := c.userRepo.GetUserByID(context.Background(), data.DriverRefID); err == nil && len(user.Phone) > 0 {
		driverPhone = user.Phone[0]
	}

	var orderOID primitive.ObjectID
	if oid, err := primitive.ObjectIDFromHex(data.OrderID); err == nil {
		orderOID = oid
	}

	debtStatus := "earned"
	description := "Trip earnings"
	if data.PaymentStatus == "debt" {
		debtStatus = "unpaid"
		description = fmt.Sprintf("Trip earnings (customer debt: Rp %.0f)", data.DebtAmount)
	}

	debt := &models.Debt{
		DriverPhone: driverPhone,
		OrderID:     orderOID,
		OrderNumber: data.TripID,
		Amount:      data.FinalPrice,
		Remaining:   driverPayout,
		Status:      debtStatus,
		Description: description,
		CreatedAt:   time.Now(),
	}
	if err := c.debtRepo.Create(context.Background(), debt); err != nil {
		log.Printf("failed to create debt/earning for driver %s: %v", driverPhone, err)
		return
	}

	c.userRepo.IncrementDriverCompletedOrder(context.Background(), data.DriverRefID)
	c.userRepo.IncrementDriverEarnings(context.Background(), data.DriverRefID, driverPayout)
	log.Printf("trip completed: driver=%s phone=%s price=%.0f payout=%.0f status=%s", data.DriverRefID, driverPhone, data.FinalPrice, driverPayout, debtStatus)
}

func (c *TripEventConsumer) handleTripCancelled(evt events.Event) {
	dataBytes, _ := json.Marshal(evt.Data)
	var data events.TripCancelledData
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return
	}
	c.userRepo.IncrementDriverFailedOrder(context.Background(), data.DriverRefID)
	log.Printf("trip cancelled: driver=%s reason=%s", data.DriverRefID, data.Reason)
}

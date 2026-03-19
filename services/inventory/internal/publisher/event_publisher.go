package publisher

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rts/inventory/internal/domain"
)

type EventPublisher struct {
	uri          string
	exchangeName string
	queueName    string
	conn         *amqp.Connection
	ch           *amqp.Channel
	mu           sync.Mutex
}

func NewEventPublisher(uri, exchangeName, queueName string) *EventPublisher {
	return &EventPublisher{
		uri:          uri,
		exchangeName: exchangeName,
		queueName:    queueName,
	}
}

func (p *EventPublisher) Connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ch != nil && !p.ch.IsClosed() {
		return nil
	}

	conn, err := amqp.Dial(p.uri)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}

	if err := ch.ExchangeDeclare(p.exchangeName, "direct", true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return err
	}

	_, err = ch.QueueDeclare(p.queueName, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return err
	}

	if err := ch.QueueBind(p.queueName, p.queueName, p.exchangeName, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return err
	}

	p.conn = conn
	p.ch = ch
	return nil
}

func (p *EventPublisher) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ch != nil {
		p.ch.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}

func (p *EventPublisher) publish(ctx context.Context, eventType string, payload interface{}) {
	event := domain.InventoryEvent{
		EventType:  eventType,
		OccurredOn: time.Now().UTC(),
		Payload:    payload,
	}

	body, err := json.Marshal(event)
	if err != nil {
		slog.Error("failed to marshal event", "type", eventType, "error", err)
		return
	}

	if err := p.Connect(); err != nil {
		slog.Warn("failed to connect to RabbitMQ for publishing", "type", eventType, "error", err)
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	err = p.ch.PublishWithContext(ctx,
		p.exchangeName,
		p.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		slog.Warn("failed to publish event", "type", eventType, "error", err)
	} else {
		slog.Info("published event", "type", eventType)
	}
}

func (p *EventPublisher) PublishStockUpdated(ctx context.Context, payload domain.StockUpdatedPayload) {
	p.publish(ctx, "inventory.stock.updated", payload)
}

func (p *EventPublisher) PublishStockLow(ctx context.Context, payload domain.StockLowPayload) {
	p.publish(ctx, "inventory.stock.low", payload)
}

func (p *EventPublisher) PublishStockOut(ctx context.Context, payload domain.StockOutPayload) {
	p.publish(ctx, "inventory.stock.out", payload)
}

func (p *EventPublisher) PublishReservationCreated(ctx context.Context, payload domain.ReservationEventPayload) {
	p.publish(ctx, "inventory.reservation.created", payload)
}

func (p *EventPublisher) PublishReservationConfirmed(ctx context.Context, payload domain.ReservationEventPayload) {
	p.publish(ctx, "inventory.reservation.confirmed", payload)
}

func (p *EventPublisher) PublishReservationReleased(ctx context.Context, payload domain.ReservationEventPayload) {
	p.publish(ctx, "inventory.reservation.released", payload)
}

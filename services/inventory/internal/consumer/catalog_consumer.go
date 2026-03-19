package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rts/inventory/internal/domain"
	"github.com/rts/inventory/internal/service"
)

const (
	catalogExchange   = "catalog.exchange"
	catalogRoutingKey = "catalog.events"
)

type CatalogConsumer struct {
	uri              string
	queueName        string
	inventoryService *service.InventoryService
	conn             *amqp.Connection
	ch               *amqp.Channel
}

func NewCatalogConsumer(uri, queueName string, inventoryService *service.InventoryService) *CatalogConsumer {
	return &CatalogConsumer{
		uri:              uri,
		queueName:        queueName,
		inventoryService: inventoryService,
	}
}

func (c *CatalogConsumer) Start(ctx context.Context) error {
	conn, err := amqp.Dial(c.uri)
	if err != nil {
		return err
	}
	c.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	c.ch = ch

	// Declare exchange (idempotent)
	if err := ch.ExchangeDeclare(catalogExchange, "direct", true, false, false, false, nil); err != nil {
		return err
	}

	// Declare our queue
	_, err = ch.QueueDeclare(c.queueName, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": c.queueName + ".dlq",
	})
	if err != nil {
		return err
	}

	// Declare dead-letter queue
	_, err = ch.QueueDeclare(c.queueName+".dlq", true, false, false, false, nil)
	if err != nil {
		return err
	}

	// Bind to catalog exchange
	if err := ch.QueueBind(c.queueName, catalogRoutingKey, catalogExchange, false, nil); err != nil {
		return err
	}

	// Set prefetch
	if err := ch.Qos(10, 0, false); err != nil {
		return err
	}

	msgs, err := ch.Consume(c.queueName, "inventory-consumer", false, false, false, false, nil)
	if err != nil {
		return err
	}

	slog.Info("catalog consumer started", "queue", c.queueName)

	go func() {
		for {
			select {
			case <-ctx.Done():
				slog.Info("catalog consumer shutting down")
				return
			case msg, ok := <-msgs:
				if !ok {
					slog.Info("catalog consumer channel closed")
					return
				}
				c.handleMessage(ctx, msg)
			}
		}
	}()

	return nil
}

func (c *CatalogConsumer) Close() {
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *CatalogConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	var envelope domain.CatalogEventEnvelope
	if err := json.Unmarshal(msg.Body, &envelope); err != nil {
		slog.Error("malformed catalog event JSON, sending to DLQ", "error", err)
		_ = msg.Nack(false, false) // NACK without requeue → dead-letter
		return
	}

	slog.Info("received catalog event", "type", envelope.EventType)

	var err error
	switch envelope.EventType {
	case "catalog.product.created":
		var payload domain.CatalogProductCreatedPayload
		if err = json.Unmarshal(envelope.Payload, &payload); err != nil {
			slog.Error("failed to unmarshal product.created payload", "error", err)
			_ = msg.Nack(false, false)
			return
		}
		err = c.inventoryService.HandleProductCreated(ctx, payload)

	case "catalog.product.updated":
		var payload domain.CatalogProductUpdatedPayload
		if err = json.Unmarshal(envelope.Payload, &payload); err != nil {
			slog.Error("failed to unmarshal product.updated payload", "error", err)
			_ = msg.Nack(false, false)
			return
		}
		err = c.inventoryService.HandleProductUpdated(ctx, payload)

	case "catalog.product.status_changed":
		var payload domain.CatalogProductStatusChangedPayload
		if err = json.Unmarshal(envelope.Payload, &payload); err != nil {
			slog.Error("failed to unmarshal product.status_changed payload", "error", err)
			_ = msg.Nack(false, false)
			return
		}
		err = c.inventoryService.HandleProductStatusChanged(ctx, payload)

	case "catalog.product.deleted":
		var payload domain.CatalogProductDeletedPayload
		if err = json.Unmarshal(envelope.Payload, &payload); err != nil {
			slog.Error("failed to unmarshal product.deleted payload", "error", err)
			_ = msg.Nack(false, false)
			return
		}
		err = c.inventoryService.HandleProductDeleted(ctx, payload)

	default:
		slog.Info("unknown catalog event type, skipping", "type", envelope.EventType)
		_ = msg.Ack(false)
		return
	}

	if err != nil {
		slog.Error("failed to handle catalog event", "type", envelope.EventType, "error", err)
		_ = msg.Nack(false, true) // Requeue for retry
		return
	}

	_ = msg.Ack(false)
	slog.Info("catalog event processed", "type", envelope.EventType)
}

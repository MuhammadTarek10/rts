using System.Text;
using System.Text.Json;
using Catalog.Api.Infrastructure.Settings;
using Catalog.Api.Shared.Abstraction;
using Microsoft.Extensions.Options;
using RabbitMQ.Client;

namespace Catalog.Api.Infrastructure.Messaging;

public sealed class RabbitMqCatalogEventPublisher(
    IOptions<RabbitMqSettings> options,
    ILogger<RabbitMqCatalogEventPublisher> logger) : ICatalogEventPublisher, IAsyncDisposable
{
    private readonly RabbitMqSettings _settings = options.Value;
    private IConnection? _connection;
    private IChannel? _channel;
    private readonly SemaphoreSlim _lock = new(1, 1);

    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNamingPolicy = JsonNamingPolicy.CamelCase,
    };

    public async Task PublishAsync(IDomainEvent domainEvent, CancellationToken cancellationToken = default)
    {
        try
        {
            var channel = await GetOrCreateChannelAsync(cancellationToken);
            var envelope = new
            {
                eventType = domainEvent.EventType,
                occurredOn = domainEvent.OccurredOnUtc,
                payload = domainEvent
            };
            var body = Encoding.UTF8.GetBytes(JsonSerializer.Serialize(envelope, JsonOptions));

            var props = new BasicProperties
            {
                Persistent = true,
                ContentType = "application/json",
            };

            await channel.BasicPublishAsync(
                exchange: _settings.ExchangeName,
                routingKey: _settings.QueueName,
                mandatory: false,
                basicProperties: props,
                body: body,
                cancellationToken: cancellationToken);

            logger.LogInformation("Published event {EventType} for catalog", domainEvent.EventType);
        }
        catch (Exception ex)
        {
            // Fire-and-forget: log warning, don't fail the request
            logger.LogWarning(ex, "Failed to publish event {EventType} to RabbitMQ — event will be lost", domainEvent.EventType);
        }
    }

    private async Task<IChannel> GetOrCreateChannelAsync(CancellationToken cancellationToken)
    {
        if (_channel is { IsOpen: true })
            return _channel;

        await _lock.WaitAsync(cancellationToken);
        try
        {
            if (_channel is { IsOpen: true })
                return _channel;

            // Dispose old resources before reconnecting
            if (_channel is not null)
            {
                await _channel.DisposeAsync();
                _channel = null;
            }
            if (_connection is not null)
            {
                await _connection.DisposeAsync();
                _connection = null;
            }

            var factory = new ConnectionFactory { Uri = new Uri(_settings.ConnectionString) };
            _connection = await factory.CreateConnectionAsync(cancellationToken);
            _channel = await _connection.CreateChannelAsync(cancellationToken: cancellationToken);

            await _channel.ExchangeDeclareAsync(_settings.ExchangeName, ExchangeType.Direct, durable: true, cancellationToken: cancellationToken);
            await _channel.QueueDeclareAsync(_settings.QueueName, durable: true, exclusive: false, autoDelete: false, cancellationToken: cancellationToken);
            await _channel.QueueBindAsync(_settings.QueueName, _settings.ExchangeName, _settings.QueueName, cancellationToken: cancellationToken);

            return _channel;
        }
        finally
        {
            _lock.Release();
        }
    }

    public async ValueTask DisposeAsync()
    {
        if (_channel is not null) await _channel.DisposeAsync();
        if (_connection is not null) await _connection.DisposeAsync();
        _lock.Dispose();
    }
}

namespace Catalog.Api.Infrastructure.Settings;

public sealed class RabbitMqSettings
{
    public string ConnectionString { get; set; } = "amqp://guest:guest@localhost:5672";
    public string ExchangeName { get; set; } = "catalog.exchange";
    public string QueueName { get; set; } = "catalog.events";
}

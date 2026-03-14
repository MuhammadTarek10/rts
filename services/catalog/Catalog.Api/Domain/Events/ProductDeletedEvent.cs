namespace Catalog.Api.Domain.Events;

public sealed record ProductDeletedEvent(
    string ProductId,
    string Sku
) : Shared.Abstraction.IDomainEvent
{
    public DateTime OccurredOnUtc { get; } = DateTime.UtcNow;
    public string EventType { get; } = "catalog.product.deleted";
}

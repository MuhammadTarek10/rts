namespace Catalog.Api.Domain.Events;

public sealed record ProductCreatedEvent(
    string ProductId,
    string Sku,
    string Title,
    string? BrandId,
    IReadOnlyList<string> CategoryIds,
    decimal Price,
    string Currency
) : Shared.Abstraction.IDomainEvent
{
    public DateTime OccurredOnUtc { get; } = DateTime.UtcNow;
    public string EventType { get; } = "catalog.product.created";
}

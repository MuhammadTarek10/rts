namespace Catalog.Api.Domain.Events;

public sealed record ProductCreatedEvent(
    string ProductId,
    string Sku,
    string Title,
    string? BrandId,
    IReadOnlyList<string> CategoryIds,
    decimal Price,
    string Currency,
    IReadOnlyList<ProductVariantEventData> Variants
) : Shared.Abstraction.IDomainEvent
{
    public DateTime OccurredOnUtc { get; } = DateTime.UtcNow;
    public string EventType { get; } = "catalog.product.created";
}

public sealed record ProductVariantEventData(
    string VariantId,
    string Sku,
    Dictionary<string, string> Attributes
);

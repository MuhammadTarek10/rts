namespace Catalog.Api.Domain.Events;

public sealed record ProductUpdatedEvent(
    string ProductId,
    IReadOnlyList<string> ChangedFields
) : Shared.Abstraction.IDomainEvent
{
    public DateTime OccurredOnUtc { get; } = DateTime.UtcNow;
    public string EventType { get; } = "catalog.product.updated";
}

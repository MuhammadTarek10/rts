namespace Catalog.Api.Domain.Events;

public sealed record ProductStatusChangedEvent(
    string ProductId,
    string OldStatus,
    string NewStatus
) : Shared.Abstraction.IDomainEvent
{
    public DateTime OccurredOnUtc { get; } = DateTime.UtcNow;
    public string EventType { get; } = "catalog.product.status_changed";
}

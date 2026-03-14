namespace Catalog.Api.Domain.Events;

public sealed record BrandCreatedEvent(
    string BrandId,
    string Name
) : Shared.Abstraction.IDomainEvent
{
    public DateTime OccurredOnUtc { get; } = DateTime.UtcNow;
    public string EventType { get; } = "catalog.brand.created";
}

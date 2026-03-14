namespace Catalog.Api.Domain.Events;

public sealed record CategoryCreatedEvent(
    string CategoryId,
    string Name,
    string? ParentId
) : Shared.Abstraction.IDomainEvent
{
    public DateTime OccurredOnUtc { get; } = DateTime.UtcNow;
    public string EventType { get; } = "catalog.category.created";
}

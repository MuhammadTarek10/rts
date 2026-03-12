namespace Catalog.Api.Shared.Abstraction;

/// <summary>
/// Represents a domain event that captures something notable that occurred in the domain.
/// </summary>
public interface IDomainEvent
{
    /// <summary>
    /// UTC timestamp of when the event occurred.
    /// </summary>
    DateTime OccurredOnUtc { get; }

    /// <summary>
    /// Discriminator string identifying the type of domain event.
    /// </summary>
    string EventType { get; }
}

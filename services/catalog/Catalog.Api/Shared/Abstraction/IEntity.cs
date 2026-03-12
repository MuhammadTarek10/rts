namespace Catalog.Api.Shared.Abstraction;

/// <summary>
/// Marker interface for domain entities with identity and audit timestamps.
/// </summary>
public interface IEntity
{
    /// <summary>
    /// Unique identifier for the entity.
    /// </summary>
    string Id { get; }

    /// <summary>
    /// UTC timestamp of when the entity was first persisted.
    /// </summary>
    DateTime CreatedAt { get; set; }

    /// <summary>
    /// UTC timestamp of the most recent modification.
    /// </summary>
    DateTime UpdatedAt { get; set; }
}

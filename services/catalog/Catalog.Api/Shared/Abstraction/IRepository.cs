namespace Catalog.Api.Shared.Abstraction;

/// <summary>
/// Generic repository contract for entity persistence operations.
/// </summary>
/// <typeparam name="TEntity">The entity type managed by this repository.</typeparam>
public interface IRepository<TEntity> where TEntity : IEntity
{
    /// <summary>
    /// Retrieves an entity by its unique identifier.
    /// </summary>
    /// <param name="id">The entity identifier.</param>
    /// <param name="cancellationToken">Cancellation token.</param>
    /// <returns>The entity if found; otherwise <c>null</c>.</returns>
    Task<TEntity?> GetByIdAsync(string id, CancellationToken cancellationToken = default);

    /// <summary>
    /// Persists a new entity to the data store.
    /// </summary>
    /// <param name="entity">The entity to create.</param>
    /// <param name="cancellationToken">Cancellation token.</param>
    Task CreateAsync(TEntity entity, CancellationToken cancellationToken = default);

    /// <summary>
    /// Replaces an existing entity in the data store.
    /// </summary>
    /// <param name="entity">The entity with updated values.</param>
    /// <param name="cancellationToken">Cancellation token.</param>
    Task UpdateAsync(TEntity entity, CancellationToken cancellationToken = default);

    /// <summary>
    /// Creates or updates an entity, setting audit timestamps automatically.
    /// </summary>
    /// <param name="entity">The entity to save.</param>
    /// <param name="isNew">Whether the entity is being created for the first time.</param>
    /// <param name="cancellationToken">Cancellation token.</param>
    async Task SaveAsync(TEntity entity, bool isNew, CancellationToken cancellationToken = default)
    {
        var now = DateTime.UtcNow;

        if (entity.CreatedAt == default)
        {
            entity.CreatedAt = now;
        }

        entity.UpdatedAt = now;

        if (isNew)
        {
            await CreateAsync(entity, cancellationToken);
            return;
        }

        await UpdateAsync(entity, cancellationToken);
    }
}

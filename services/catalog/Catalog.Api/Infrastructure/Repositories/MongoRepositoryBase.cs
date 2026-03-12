using Catalog.Api.Shared.Abstraction;

namespace Catalog.Api.Infrastructure.Repositories;

/// <summary>
/// Base class for MongoDB-backed repositories providing structured error logging.
/// </summary>
/// <typeparam name="TEntity">The entity type managed by the repository.</typeparam>
public abstract class MongoRepositoryBase<TEntity>(ILogger logger) where TEntity : IEntity
{
    /// <summary>
    /// Executes a database operation, logging any exception before re-throwing.
    /// </summary>
    /// <typeparam name="T">The return type of the operation.</typeparam>
    /// <param name="operation">The async database operation to execute.</param>
    /// <param name="errorMessage">Structured log message template on failure.</param>
    /// <param name="args">Arguments for the log message template.</param>
    protected async Task<T> ExecuteAsync<T>(Func<Task<T>> operation, string errorMessage, params object?[] args)
    {
        try
        {
            return await operation().ConfigureAwait(false);
        }
        catch (Exception ex)
        {
            logger.LogError(ex, errorMessage, args);
            throw;
        }
    }

    /// <summary>
    /// Executes a void database operation, logging any exception before re-throwing.
    /// </summary>
    /// <param name="operation">The async database operation to execute.</param>
    /// <param name="errorMessage">Structured log message template on failure.</param>
    /// <param name="args">Arguments for the log message template.</param>
    protected async Task ExecuteAsync(Func<Task> operation, string errorMessage, params object?[] args)
    {
        try
        {
            await operation().ConfigureAwait(false);
        }
        catch (Exception ex)
        {
            logger.LogError(ex, errorMessage, args);
            throw;
        }
    }
}

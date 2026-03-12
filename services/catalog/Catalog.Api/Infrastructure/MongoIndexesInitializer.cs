using Catalog.Api.Domain.Entities;
using MongoDB.Driver;

namespace Catalog.Api.Infrastructure;

/// <summary>
/// Hosted service that ensures required MongoDB indexes exist on application startup.
/// </summary>
public sealed class MongoIndexesInitializer(IMongoDatabase database, ILogger<MongoIndexesInitializer> logger) : IHostedService
{
    /// <summary>
    /// Creates unique and compound indexes for the products collection.
    /// </summary>
    public async Task StartAsync(CancellationToken cancellationToken)
    {
        var products = database.GetCollection<Product>("products");

        var keys = Builders<Product>.IndexKeys;
        var indexModels = new[]
        {
            new CreateIndexModel<Product>(keys.Ascending(product => product.Sku), new CreateIndexOptions { Unique = true }),
            new CreateIndexModel<Product>(keys.Ascending(product => product.Slug), new CreateIndexOptions { Unique = true }),
            new CreateIndexModel<Product>(keys.Ascending(product => product.Status).Ascending(product => product.BrandId)),
        };

        await products.Indexes.CreateManyAsync(indexModels, cancellationToken);

        logger.LogInformation("MongoDB indexes for products collection created successfully");
    }

    /// <inheritdoc />
    public Task StopAsync(CancellationToken cancellationToken)
    {
        return Task.CompletedTask;
    }
}

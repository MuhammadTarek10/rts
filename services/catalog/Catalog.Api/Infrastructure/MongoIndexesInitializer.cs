using Catalog.Api.Domain.Entities;
using MongoDB.Bson;
using MongoDB.Driver;

namespace Catalog.Api.Infrastructure;

/// <summary>
/// Hosted service that ensures required MongoDB indexes exist on application startup.
/// </summary>
public sealed class MongoIndexesInitializer(IMongoDatabase database, ILogger<MongoIndexesInitializer> logger) : IHostedService
{
    /// <inheritdoc />
    public async Task StartAsync(CancellationToken cancellationToken)
    {
        await CreateProductIndexesAsync(cancellationToken);
        await CreateCategoryIndexesAsync(cancellationToken);
        await CreateBrandIndexesAsync(cancellationToken);
        logger.LogInformation("MongoDB indexes created successfully for all collections");
    }

    /// <inheritdoc />
    public Task StopAsync(CancellationToken cancellationToken) => Task.CompletedTask;

    private async Task CreateProductIndexesAsync(CancellationToken cancellationToken)
    {
        var products = database.GetCollection<Product>("products");
        var keys = Builders<Product>.IndexKeys;

        var indexModels = new[]
        {
            new CreateIndexModel<Product>(keys.Ascending(p => p.Sku), new CreateIndexOptions { Unique = true }),
            new CreateIndexModel<Product>(keys.Ascending(p => p.Slug), new CreateIndexOptions { Unique = true }),
            new CreateIndexModel<Product>(keys.Ascending(p => p.Status).Ascending(p => p.BrandId)),
            new CreateIndexModel<Product>(
                new BsonDocument
                {
                    { "title", "text" },
                    { "description", "text" },
                },
                new CreateIndexOptions
                {
                    Weights = new BsonDocument
                    {
                        { "title", 10 },
                        { "description", 1 },
                    },
                    DefaultLanguage = "english",
                }),
            new CreateIndexModel<Product>(keys.Ascending(p => p.CategoryIds)),
            new CreateIndexModel<Product>(keys.Ascending(p => p.Price.Amount)),
        };

        await products.Indexes.CreateManyAsync(indexModels, cancellationToken);
    }

    private async Task CreateCategoryIndexesAsync(CancellationToken cancellationToken)
    {
        var categories = database.GetCollection<Category>("categories");
        var keys = Builders<Category>.IndexKeys;

        var indexModels = new[]
        {
            new CreateIndexModel<Category>(keys.Ascending(c => c.Slug), new CreateIndexOptions { Unique = true }),
            new CreateIndexModel<Category>(keys.Ascending(c => c.ParentId)),
            new CreateIndexModel<Category>(keys.Ascending(c => c.Path)),
        };

        await categories.Indexes.CreateManyAsync(indexModels, cancellationToken);
    }

    private async Task CreateBrandIndexesAsync(CancellationToken cancellationToken)
    {
        var brands = database.GetCollection<Brand>("brands");
        var keys = Builders<Brand>.IndexKeys;

        var indexModels = new[]
        {
            new CreateIndexModel<Brand>(keys.Ascending(b => b.Slug), new CreateIndexOptions { Unique = true }),
            new CreateIndexModel<Brand>(keys.Ascending(b => b.Name)),
        };

        await brands.Indexes.CreateManyAsync(indexModels, cancellationToken);
    }
}

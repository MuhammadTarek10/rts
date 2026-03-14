using Catalog.Api.Domain.Entities;
using Catalog.Api.Domain.Interfaces;
using Catalog.Api.Domain.Models;
using Catalog.Api.Shared.Exceptions;
using MongoDB.Bson;
using MongoDB.Bson.Serialization;
using MongoDB.Driver;

namespace Catalog.Api.Infrastructure.Repositories;

/// <summary>
/// MongoDB-backed repository for <see cref="Product"/> aggregate persistence.
/// </summary>
public sealed class ProductRepository(IMongoDatabase database, ILogger<ProductRepository> logger)
    : MongoRepositoryBase<Product>(logger), IProductRepository
{
    private readonly IMongoCollection<Product> _products = database.GetCollection<Product>("products");

    /// <inheritdoc />
    public Task<Product?> GetByIdAsync(string id, CancellationToken cancellationToken = default) =>
        ExecuteAsync<Product?>(
            () => _products.Find(p => p.Id == id).FirstOrDefaultAsync(cancellationToken)!,
            "Error retrieving product {ProductId}", id);

    /// <inheritdoc />
    public Task CreateAsync(Product product, CancellationToken cancellationToken = default) =>
        ExecuteAsync(
            () => _products.InsertOneAsync(product, cancellationToken: cancellationToken),
            "Error creating product {ProductId}", product.Id);

    /// <inheritdoc />
    public Task UpdateAsync(Product product, CancellationToken cancellationToken = default) =>
        ExecuteAsync(
            () => _products.ReplaceOneAsync(
                existing => existing.Id == product.Id,
                product,
                cancellationToken: cancellationToken),
            "Error updating product {ProductId}", product.Id);

    /// <inheritdoc />
    public Task<bool> ExistsBySkuAsync(string sku, CancellationToken cancellationToken = default) =>
        ExecuteAsync(
            () => _products.Find(p => p.Sku == sku).AnyAsync(cancellationToken),
            "Error checking SKU existence {Sku}", sku);

    /// <inheritdoc />
    public Task<bool> ExistsBySlugAsync(string slug, CancellationToken cancellationToken = default) =>
        ExecuteAsync(
            () => _products.Find(p => p.Slug == slug).AnyAsync(cancellationToken),
            "Error checking slug existence {Slug}", slug);

    /// <inheritdoc />
    public Task<Product?> GetBySlugAsync(string slug, CancellationToken cancellationToken = default) =>
        ExecuteAsync<Product?>(
            () => _products.Find(p => p.Slug == slug).FirstOrDefaultAsync(cancellationToken)!,
            "Error retrieving product by slug {Slug}", slug);

    /// <inheritdoc />
    public Task ReplaceIfVersionMatchesAsync(Product product, int expectedVersion, CancellationToken cancellationToken = default) =>
        ExecuteAsync(async () =>
        {
            var filter = Builders<Product>.Filter.And(
                Builders<Product>.Filter.Eq(p => p.Id, product.Id),
                Builders<Product>.Filter.Eq(p => p.Version, expectedVersion));
            var result = await _products.ReplaceOneAsync(filter, product, cancellationToken: cancellationToken);
            if (result.ModifiedCount == 0)
            {
                var exists = await _products.Find(p => p.Id == product.Id).AnyAsync(cancellationToken);
                if (exists)
                    throw new ConflictException("VERSION_CONFLICT", "Product was modified by another user. Please refresh and try again.");
            }
        }, "Error replacing product with version check {ProductId}", product.Id);

    /// <inheritdoc />
    public Task<bool> HasCategoryAsync(string categoryId, CancellationToken cancellationToken = default) =>
        ExecuteAsync(
            () => _products.Find(p => p.CategoryIds.Contains(categoryId)).AnyAsync(cancellationToken),
            "Error checking category usage {CategoryId}", categoryId);

    /// <inheritdoc />
    public Task<bool> HasBrandAsync(string brandId, CancellationToken cancellationToken = default) =>
        ExecuteAsync(
            () => _products.Find(p => p.BrandId == brandId).AnyAsync(cancellationToken),
            "Error checking brand usage {BrandId}", brandId);

    /// <inheritdoc />
    public async Task<(IReadOnlyList<Product> Items, long TotalCount)> SearchAsync(
        ProductSearchCriteria criteria,
        CancellationToken cancellationToken = default)
    {
        return await ExecuteAsync(async () =>
        {
            var filterDef = BuildSearchFilter(criteria);
            var skip = (criteria.Page - 1) * criteria.PageSize;
            var limit = criteria.PageSize;

            var pipeline = new[]
            {
                new BsonDocument("$match", filterDef.Render(
                    BsonSerializer.SerializerRegistry.GetSerializer<Product>(),
                    BsonSerializer.SerializerRegistry)),
                new BsonDocument("$facet", new BsonDocument
                {
                    {
                        "data", new BsonArray
                        {
                            new BsonDocument("$sort", GetSortDocument(criteria.SortBy)),
                            new BsonDocument("$skip", skip),
                            new BsonDocument("$limit", limit),
                        }
                    },
                    { "count", new BsonArray { new BsonDocument("$count", "total") } }
                })
            };

            var aggregateOptions = new AggregateOptions { AllowDiskUse = true };
            var cursor = await _products.AggregateAsync<BsonDocument>(pipeline, aggregateOptions, cancellationToken);
            var result = await cursor.FirstOrDefaultAsync(cancellationToken);

            if (result is null) return (Array.Empty<Product>(), 0L);

            var dataArray = result["data"].AsBsonArray;
            var countArray = result["count"].AsBsonArray;

            var items = dataArray
                .Select(doc => BsonSerializer.Deserialize<Product>(doc.AsBsonDocument))
                .ToList();
            var total = countArray.Count > 0 ? countArray[0]["total"].ToInt64() : 0L;

            return ((IReadOnlyList<Product>)items, total);
        }, "Error searching products");
    }

    private FilterDefinition<Product> BuildSearchFilter(ProductSearchCriteria criteria)
    {
        var filters = new List<FilterDefinition<Product>>();

        if (!string.IsNullOrWhiteSpace(criteria.Query))
            filters.Add(Builders<Product>.Filter.Text(criteria.Query, new TextSearchOptions { CaseSensitive = false }));

        if (criteria.CategoryIds is { Count: > 0 })
            filters.Add(Builders<Product>.Filter.AnyIn(p => p.CategoryIds, criteria.CategoryIds));
        else if (!string.IsNullOrWhiteSpace(criteria.CategoryId))
            filters.Add(Builders<Product>.Filter.AnyEq(p => p.CategoryIds, criteria.CategoryId));

        if (!string.IsNullOrWhiteSpace(criteria.BrandId))
            filters.Add(Builders<Product>.Filter.Eq(p => p.BrandId, criteria.BrandId));

        if (criteria.MinPrice.HasValue)
            filters.Add(Builders<Product>.Filter.Gte(p => p.Price.Amount, criteria.MinPrice.Value));

        if (criteria.MaxPrice.HasValue)
            filters.Add(Builders<Product>.Filter.Lte(p => p.Price.Amount, criteria.MaxPrice.Value));

        if (criteria.Status.HasValue)
            filters.Add(Builders<Product>.Filter.Eq(p => p.Status, criteria.Status.Value));

        return filters.Count == 0
            ? Builders<Product>.Filter.Empty
            : Builders<Product>.Filter.And(filters);
    }

    private static BsonDocument GetSortDocument(string? sortBy) => sortBy switch
    {
        "price_asc" => new BsonDocument("price.amount", 1),
        "price_desc" => new BsonDocument("price.amount", -1),
        "relevance" => new BsonDocument("score", new BsonDocument("$meta", "textScore")),
        _ => new BsonDocument("createdAt", -1),
    };
}

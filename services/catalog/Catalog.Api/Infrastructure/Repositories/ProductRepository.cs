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
}

using Catalog.Api.Domain.Entities;
using Catalog.Api.Domain.Interfaces;
using MongoDB.Driver;

namespace Catalog.Api.Infrastructure.Repositories;

/// <summary>
/// MongoDB-backed repository for <see cref="Brand"/> persistence.
/// </summary>
public sealed class BrandRepository(IMongoDatabase database, ILogger<BrandRepository> logger)
    : MongoRepositoryBase<Brand>(logger), IBrandRepository
{
    private readonly IMongoCollection<Brand> _brands = database.GetCollection<Brand>("brands");
    private readonly IMongoCollection<Product> _products = database.GetCollection<Product>("products");

    /// <inheritdoc />
    public Task<Brand?> GetByIdAsync(string id, CancellationToken cancellationToken = default) =>
        ExecuteAsync<Brand?>(
            () => _brands.Find(b => b.Id == id).FirstOrDefaultAsync(cancellationToken)!,
            "Error retrieving brand {BrandId}", id);

    /// <inheritdoc />
    public Task CreateAsync(Brand brand, CancellationToken cancellationToken = default) =>
        ExecuteAsync(
            () => _brands.InsertOneAsync(brand, cancellationToken: cancellationToken),
            "Error creating brand {BrandId}", brand.Id);

    /// <inheritdoc />
    public Task UpdateAsync(Brand brand, CancellationToken cancellationToken = default) =>
        ExecuteAsync(
            () => _brands.ReplaceOneAsync(
                existing => existing.Id == brand.Id,
                brand,
                cancellationToken: cancellationToken),
            "Error updating brand {BrandId}", brand.Id);

    /// <inheritdoc />
    public Task<bool> ExistsBySlugAsync(string slug, CancellationToken cancellationToken = default) =>
        ExecuteAsync(
            () => _brands.Find(b => b.Slug == slug).AnyAsync(cancellationToken),
            "Error checking slug existence {Slug}", slug);

    /// <inheritdoc />
    public Task<bool> ExistsAsync(string id, CancellationToken cancellationToken = default) =>
        ExecuteAsync(
            () => _brands.Find(b => b.Id == id).AnyAsync(cancellationToken),
            "Error checking brand existence {BrandId}", id);

    /// <inheritdoc />
    public Task<IReadOnlyList<Brand>> GetAllActiveAsync(CancellationToken cancellationToken = default) =>
        ExecuteAsync<IReadOnlyList<Brand>>(
            async () => await _brands.Find(b => b.IsActive).ToListAsync(cancellationToken),
            "Error retrieving all active brands");

    /// <inheritdoc />
    public Task<bool> HasProductsAsync(string brandId, CancellationToken cancellationToken = default) =>
        ExecuteAsync(
            () => _products.Find(p => p.BrandId == brandId).AnyAsync(cancellationToken),
            "Error checking product usage for brand {BrandId}", brandId);

    /// <inheritdoc />
    public Task DeleteAsync(string id, CancellationToken cancellationToken = default) =>
        ExecuteAsync(
            () => _brands.DeleteOneAsync(b => b.Id == id, cancellationToken),
            "Error deleting brand {BrandId}", id);
}

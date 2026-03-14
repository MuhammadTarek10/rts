using Catalog.Api.Domain.Entities;
using Catalog.Api.Domain.Models;
using Catalog.Api.Shared.Abstraction;

namespace Catalog.Api.Domain.Interfaces;

/// <summary>
/// Repository contract for <see cref="Product"/> persistence with domain-specific queries.
/// </summary>
public interface IProductRepository : IRepository<Product>
{
    /// <summary>
    /// Checks whether a product with the given SKU already exists.
    /// </summary>
    /// <param name="sku">The SKU to check.</param>
    /// <param name="cancellationToken">Cancellation token.</param>
    /// <returns><c>true</c> if a product with the SKU exists; otherwise <c>false</c>.</returns>
    Task<bool> ExistsBySkuAsync(string sku, CancellationToken cancellationToken = default);

    /// <summary>
    /// Checks whether a product with the given slug already exists.
    /// </summary>
    /// <param name="slug">The slug to check.</param>
    /// <param name="cancellationToken">Cancellation token.</param>
    /// <returns><c>true</c> if a product with the slug exists; otherwise <c>false</c>.</returns>
    Task<bool> ExistsBySlugAsync(string slug, CancellationToken cancellationToken = default);

    Task<Product?> GetBySlugAsync(string slug, CancellationToken cancellationToken = default);
    Task ReplaceIfVersionMatchesAsync(Product product, int expectedVersion, CancellationToken cancellationToken = default);
    Task<(IReadOnlyList<Product> Items, long TotalCount)> SearchAsync(ProductSearchCriteria criteria, CancellationToken cancellationToken = default);
    Task<bool> HasCategoryAsync(string categoryId, CancellationToken cancellationToken = default);
    Task<bool> HasBrandAsync(string brandId, CancellationToken cancellationToken = default);
}

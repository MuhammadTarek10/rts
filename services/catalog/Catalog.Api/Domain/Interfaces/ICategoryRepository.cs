using Catalog.Api.Domain.Entities;
using Catalog.Api.Shared.Abstraction;

namespace Catalog.Api.Domain.Interfaces;

public interface ICategoryRepository : IRepository<Category>
{
    Task<bool> ExistsBySlugAsync(string slug, CancellationToken cancellationToken = default);
    Task<Category?> GetBySlugAsync(string slug, CancellationToken cancellationToken = default);
    Task<IReadOnlyList<Category>> GetAllActiveAsync(CancellationToken cancellationToken = default);
    Task<IReadOnlyList<Category>> GetChildrenAsync(string parentId, CancellationToken cancellationToken = default);
    Task<bool> HasProductsAsync(string categoryId, CancellationToken cancellationToken = default);
    Task<bool> HasChildrenAsync(string categoryId, CancellationToken cancellationToken = default);
    // Returns all descendant category IDs (for search expansion)
    Task<IReadOnlyList<string>> GetDescendantIdsAsync(string categoryId, CancellationToken cancellationToken = default);
    Task DeleteAsync(string id, CancellationToken cancellationToken = default);
}

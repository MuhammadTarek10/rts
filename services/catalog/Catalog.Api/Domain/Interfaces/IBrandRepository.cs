using Catalog.Api.Domain.Entities;
using Catalog.Api.Shared.Abstraction;

namespace Catalog.Api.Domain.Interfaces;

public interface IBrandRepository : IRepository<Brand>
{
    Task<bool> ExistsBySlugAsync(string slug, CancellationToken cancellationToken = default);
    Task<bool> ExistsAsync(string id, CancellationToken cancellationToken = default);
    Task<IReadOnlyList<Brand>> GetAllActiveAsync(CancellationToken cancellationToken = default);
    Task<bool> HasProductsAsync(string brandId, CancellationToken cancellationToken = default);
    Task DeleteAsync(string id, CancellationToken cancellationToken = default);
}

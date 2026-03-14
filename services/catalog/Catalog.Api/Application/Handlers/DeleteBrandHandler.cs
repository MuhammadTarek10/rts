using Catalog.Api.Infrastructure;
using Catalog.Api.Infrastructure.Services;
using Catalog.Api.Shared.Exceptions;

namespace Catalog.Api.Application.Handlers;

public sealed class DeleteBrandHandler(
    CatalogUnitOfWork uow,
    ICacheService cache,
    ILogger<DeleteBrandHandler> logger)
{
    public async Task HandleAsync(string id, CancellationToken cancellationToken = default)
    {
        logger.LogInformation("Deleting brand {BrandId}", id);

        var brand = await uow.Brands.GetByIdAsync(id, cancellationToken)
            ?? throw new NotFoundException("BRAND_NOT_FOUND", $"Brand '{id}' not found.");

        if (await uow.Brands.HasProductsAsync(id, cancellationToken))
            throw new ConflictException("BRAND_HAS_PRODUCTS", "Cannot delete brand: products reference it.");

        await uow.Brands.DeleteAsync(id, cancellationToken);
        await cache.RemoveAsync("catalog:brands:all", cancellationToken);

        logger.LogInformation("Brand {BrandId} deleted", id);
    }
}

using Catalog.Api.Application.DTOs;
using Catalog.Api.Infrastructure;
using Catalog.Api.Infrastructure.Services;
using Catalog.Api.Shared.Exceptions;

namespace Catalog.Api.Application.Handlers;

public sealed class UpdateBrandHandler(
    CatalogUnitOfWork uow,
    ICacheService cache,
    ILogger<UpdateBrandHandler> logger)
{
    public async Task<BrandResponseDto> HandleAsync(
        string id,
        UpdateBrandDto request,
        CancellationToken cancellationToken = default)
    {
        logger.LogInformation("Updating brand {BrandId}", id);

        var brand = await uow.Brands.GetByIdAsync(id, cancellationToken)
            ?? throw new NotFoundException("BRAND_NOT_FOUND", $"Brand '{id}' not found.");

        brand.Name = request.Name.Trim();
        brand.Description = request.Description?.Trim();
        brand.LogoUrl = request.LogoUrl;
        brand.Website = request.Website;
        brand.IsActive = request.IsActive;
        brand.UpdatedAt = DateTime.UtcNow;

        await uow.Brands.UpdateAsync(brand, cancellationToken);
        await cache.RemoveAsync("catalog:brands:all", cancellationToken);

        logger.LogInformation("Brand {BrandId} updated", id);
        return BrandResponseDto.FromEntity(brand);
    }
}

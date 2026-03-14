using Catalog.Api.Application.DTOs;
using Catalog.Api.Infrastructure;
using Catalog.Api.Infrastructure.Services;
using Catalog.Api.Shared.Exceptions;

namespace Catalog.Api.Application.Handlers;

public sealed class GetProductBySlugHandler(
    CatalogUnitOfWork uow,
    ICacheService cache,
    ILogger<GetProductBySlugHandler> logger)
{
    public async Task<ProductResponseDto> HandleAsync(string slug, CancellationToken cancellationToken = default)
    {
        var cacheKey = $"catalog:product:slug:{slug}";
        var cached = await cache.GetAsync<ProductResponseDto>(cacheKey, cancellationToken);
        if (cached is not null) return cached;

        var product = await uow.Products.GetBySlugAsync(slug, cancellationToken)
            ?? throw new NotFoundException("PRODUCT_NOT_FOUND", $"Product with slug '{slug}' not found.");

        var dto = ProductResponseDto.FromEntity(product);
        await cache.SetAsync(cacheKey, dto, TimeSpan.FromMinutes(5), cancellationToken);
        return dto;
    }
}

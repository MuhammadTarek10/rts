using Catalog.Api.Application.DTOs;
using Catalog.Api.Infrastructure;
using Catalog.Api.Infrastructure.Services;
using Catalog.Api.Shared.Exceptions;

namespace Catalog.Api.Application.Handlers;

public sealed class GetProductByIdHandler(
    CatalogUnitOfWork uow,
    ICacheService cache,
    ILogger<GetProductByIdHandler> logger)
{
    public async Task<ProductResponseDto> HandleAsync(string id, CancellationToken cancellationToken = default)
    {
        var cacheKey = $"catalog:product:{id}";
        var cached = await cache.GetAsync<ProductResponseDto>(cacheKey, cancellationToken);
        if (cached is not null) return cached;

        var product = await uow.Products.GetByIdAsync(id, cancellationToken)
            ?? throw new NotFoundException("PRODUCT_NOT_FOUND", $"Product '{id}' not found.");

        var dto = ProductResponseDto.FromEntity(product);
        await cache.SetAsync(cacheKey, dto, TimeSpan.FromMinutes(5), cancellationToken);
        return dto;
    }
}

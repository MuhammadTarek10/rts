using Catalog.Api.Application.DTOs;
using Catalog.Api.Domain.Entities;
using Catalog.Api.Infrastructure;
using Catalog.Api.Infrastructure.Services;
using Catalog.Api.Shared.Exceptions;

namespace Catalog.Api.Application.Handlers;

public sealed class ReorderProductImagesHandler(
    CatalogUnitOfWork uow,
    ICacheService cache,
    ILogger<ReorderProductImagesHandler> logger)
{
    public async Task<ProductResponseDto> HandleAsync(
        string productId,
        ReorderImagesDto request,
        CancellationToken cancellationToken = default)
    {
        logger.LogInformation("Reordering images for product {ProductId}", productId);

        var product = await uow.Products.GetByIdAsync(productId, cancellationToken)
            ?? throw new NotFoundException("PRODUCT_NOT_FOUND", $"Product '{productId}' not found.");

        var orderMap = request.Order.ToDictionary(o => o.ImageId, o => o.SortOrder);

        product.Images = product.Images.Select(img =>
        {
            if (orderMap.TryGetValue(img.ImageId, out var newOrder) && newOrder != img.SortOrder)
                return new ProductImage { ImageId = img.ImageId, Url = img.Url, Key = img.Key, AltText = img.AltText, SortOrder = newOrder, IsPrimary = img.IsPrimary };
            return img;
        }).OrderBy(img => img.SortOrder).ToList();

        product.Touch();
        await uow.Products.UpdateAsync(product, cancellationToken);

        await cache.RemoveAsync($"catalog:product:{productId}", cancellationToken);
        await cache.RemoveAsync($"catalog:product:slug:{product.Slug}", cancellationToken);

        return ProductResponseDto.FromEntity(product);
    }
}

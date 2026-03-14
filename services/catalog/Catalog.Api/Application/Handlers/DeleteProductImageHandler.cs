using Catalog.Api.Application.DTOs;
using Catalog.Api.Domain.Entities;
using Catalog.Api.Infrastructure;
using Catalog.Api.Infrastructure.Services;
using Catalog.Api.Shared.Exceptions;

namespace Catalog.Api.Application.Handlers;

public sealed class DeleteProductImageHandler(
    CatalogUnitOfWork uow,
    IImageStorageService imageStorage,
    ICacheService cache,
    ILogger<DeleteProductImageHandler> logger)
{
    public async Task<ProductResponseDto> HandleAsync(
        string productId,
        string imageId,
        CancellationToken cancellationToken = default)
    {
        logger.LogInformation("Deleting image {ImageId} from product {ProductId}", imageId, productId);

        var product = await uow.Products.GetByIdAsync(productId, cancellationToken)
            ?? throw new NotFoundException("PRODUCT_NOT_FOUND", $"Product '{productId}' not found.");

        var image = product.Images.FirstOrDefault(i => i.ImageId == imageId)
            ?? throw new NotFoundException("IMAGE_NOT_FOUND", $"Image '{imageId}' not found on product '{productId}'.");

        var remaining = product.Images.Where(i => i.ImageId != imageId).ToList();

        // If deleted image was primary and others exist, promote the lowest-sortorder image
        if (image.IsPrimary && remaining.Count > 0)
        {
            var firstImg = remaining.OrderBy(i => i.SortOrder).First();
            remaining = remaining.Select(img =>
                img.ImageId == firstImg.ImageId
                    ? new ProductImage { ImageId = img.ImageId, Url = img.Url, Key = img.Key, AltText = img.AltText, SortOrder = img.SortOrder, IsPrimary = true }
                    : img
            ).ToList();
        }

        product.Images = remaining;
        product.Touch();
        await uow.Products.UpdateAsync(product, cancellationToken);

        // Delete from MinIO after DB is updated — a MinIO orphan is recoverable,
        // but a broken DB reference (deleted file still in document) is not.
        await imageStorage.DeleteAsync(image.Key, cancellationToken);

        await cache.RemoveAsync($"catalog:product:{productId}", cancellationToken);
        await cache.RemoveAsync($"catalog:product:slug:{product.Slug}", cancellationToken);

        logger.LogInformation("Image {ImageId} deleted from product {ProductId}", imageId, productId);
        return ProductResponseDto.FromEntity(product);
    }
}

using Catalog.Api.Application.DTOs;
using Catalog.Api.Domain.Entities;
using Catalog.Api.Infrastructure;
using Catalog.Api.Infrastructure.Services;
using Catalog.Api.Shared.Exceptions;

namespace Catalog.Api.Application.Handlers;

public sealed class UploadProductImageHandler(
    CatalogUnitOfWork uow,
    IImageStorageService imageStorage,
    ICacheService cache,
    ILogger<UploadProductImageHandler> logger)
{
    private static readonly string[] AllowedContentTypes = ["image/jpeg", "image/jpg", "image/png", "image/webp"];
    private const long MaxFileSizeBytes = 5 * 1024 * 1024; // 5MB

    // Magic bytes for file type validation
    private static readonly byte[] JpegMagic = [0xFF, 0xD8, 0xFF];
    private static readonly byte[] PngMagic = [0x89, 0x50, 0x4E, 0x47];
    private static readonly byte[] WebpMagic = [0x52, 0x49, 0x46, 0x46]; // RIFF header

    public async Task<ProductResponseDto> HandleAsync(
        string productId,
        IFormFile file,
        string? altText,
        CancellationToken cancellationToken = default)
    {
        logger.LogInformation("Uploading image for product {ProductId}", productId);

        if (file.Length > MaxFileSizeBytes)
            throw new DomainException("FILE_TOO_LARGE", "Image must be 5MB or smaller.");

        var contentType = file.ContentType.ToLowerInvariant();
        if (!AllowedContentTypes.Contains(contentType))
            throw new DomainException("INVALID_CONTENT_TYPE", "Only JPEG, PNG, and WebP images are allowed.");

        var product = await uow.Products.GetByIdAsync(productId, cancellationToken)
            ?? throw new NotFoundException("PRODUCT_NOT_FOUND", $"Product '{productId}' not found.");

        if (product.Status == ProductStatus.Archived)
            throw new DomainException("PRODUCT_ARCHIVED", "Cannot upload images to an archived product.");

        // Validate magic bytes
        using var stream = file.OpenReadStream();
        await ValidateMagicBytesAsync(stream, contentType);
        stream.Seek(0, SeekOrigin.Begin);

        var (url, key) = await imageStorage.UploadAsync(stream, file.FileName, contentType, productId, cancellationToken);

        var sortOrder = product.Images.Count > 0 ? product.Images.Max(i => i.SortOrder) + 1 : 0;
        var isPrimary = product.Images.Count == 0;

        var image = new ProductImage
        {
            Url = url,
            Key = key,
            AltText = altText,
            SortOrder = sortOrder,
            IsPrimary = isPrimary,
        };

        product.Images = [..product.Images, image];
        product.Touch();
        await uow.Products.UpdateAsync(product, cancellationToken);

        await cache.RemoveAsync($"catalog:product:{productId}", cancellationToken);
        await cache.RemoveAsync($"catalog:product:slug:{product.Slug}", cancellationToken);

        logger.LogInformation("Image uploaded for product {ProductId}: {Key}", productId, key);
        return ProductResponseDto.FromEntity(product);
    }

    private static async Task ValidateMagicBytesAsync(Stream stream, string contentType)
    {
        var buffer = new byte[4];
        var read = await stream.ReadAsync(buffer, 0, 4);
        if (read < 3) throw new DomainException("INVALID_FILE", "File is too small to be a valid image.");

        var isValid = contentType switch
        {
            "image/jpeg" or "image/jpg" => buffer[..3].SequenceEqual(JpegMagic),
            "image/png" => buffer[..4].SequenceEqual(PngMagic),
            "image/webp" => buffer[..4].SequenceEqual(WebpMagic),
            _ => false
        };

        if (!isValid)
            throw new DomainException("MAGIC_BYTES_MISMATCH", "File content does not match the declared content type.");
    }
}

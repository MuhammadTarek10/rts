using Catalog.Api.Application.DTOs;
using Catalog.Api.Domain.Entities;
using Catalog.Api.Domain.Interfaces;
using Catalog.Api.Shared.Exceptions;

namespace Catalog.Api.Application.Handlers;

/// <summary>
/// Handles creation of new products, enforcing SKU and slug uniqueness.
/// </summary>
public sealed class CreateProductHandler(IProductRepository productRepository, ILogger<CreateProductHandler> logger)
{
    /// <summary>
    /// Creates a new product from the given request.
    /// </summary>
    /// <param name="request">Product creation data.</param>
    /// <param name="cancellationToken">Cancellation token.</param>
    /// <returns>The created product mapped to a response DTO.</returns>
    /// <exception cref="ConflictException">Thrown when a product with the same SKU or slug already exists.</exception>
    public async Task<ProductResponseDto> HandleAsync(
        CreateProductDto request,
        CancellationToken cancellationToken = default)
    {
        if (await productRepository.ExistsBySkuAsync(request.Sku, cancellationToken))
        {
            logger.LogWarning("Product creation conflict: SKU {Sku} already exists", request.Sku);
            throw new ConflictException("DUPLICATE_SKU", "A product with the same SKU already exists.");
        }

        if (await productRepository.ExistsBySlugAsync(request.Slug, cancellationToken))
        {
            logger.LogWarning("Product creation conflict: slug {Slug} already exists", request.Slug);
            throw new ConflictException("DUPLICATE_SLUG", "A product with the same slug already exists.");
        }

        var product = new Product
        {
            Sku = request.Sku.Trim(),
            Slug = request.Slug.Trim().ToLowerInvariant(),
            Title = request.Title.Trim(),
            Description = request.Description?.Trim(),
            BrandId = request.BrandId?.Trim(),
            CategoryIds = request.CategoryIds,
            Price = new Money
            {
                Amount = request.Amount,
                Currency = request.Currency.Trim().ToUpperInvariant(),
            },
        };

        await productRepository.SaveAsync(product, isNew: true, cancellationToken);

        logger.LogInformation("Product {ProductId} created with SKU {Sku}", product.Id, product.Sku);

        return Map(product);
    }

    /// <summary>
    /// Maps a <see cref="Product"/> entity to a <see cref="ProductResponseDto"/>.
    /// </summary>
    public static ProductResponseDto Map(Product product)
    {
        return new ProductResponseDto(
            product.Id,
            product.Sku,
            product.Slug,
            product.Title,
            product.Description,
            product.Price.Amount,
            product.Price.Currency,
            product.Status.ToString(),
            product.BrandId,
            product.CategoryIds,
            product.CreatedAt,
            product.UpdatedAt,
            product.Version
        );
    }
}

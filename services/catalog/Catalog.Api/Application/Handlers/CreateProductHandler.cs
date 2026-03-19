using Catalog.Api.Application.DTOs;
using Catalog.Api.Domain.Entities;
using Catalog.Api.Domain.Events;
using Catalog.Api.Infrastructure;
using Catalog.Api.Infrastructure.Messaging;
using Catalog.Api.Shared.Exceptions;

namespace Catalog.Api.Application.Handlers;

public sealed class CreateProductHandler(
    CatalogUnitOfWork uow,
    ICatalogEventPublisher eventPublisher,
    ILogger<CreateProductHandler> logger)
{
    public async Task<ProductResponseDto> HandleAsync(
        CreateProductDto request,
        CancellationToken cancellationToken = default)
    {
        var sw = System.Diagnostics.Stopwatch.StartNew();
        logger.LogInformation("Creating product {Sku}", request.Sku);

        if (await uow.Products.ExistsBySkuAsync(request.Sku, cancellationToken))
        {
            logger.LogWarning("Product creation conflict: SKU {Sku} already exists", request.Sku);
            throw new ConflictException("DUPLICATE_SKU", "A product with the same SKU already exists.");
        }

        if (await uow.Products.ExistsBySlugAsync(request.Slug, cancellationToken))
        {
            logger.LogWarning("Product creation conflict: slug {Slug} already exists", request.Slug);
            throw new ConflictException("DUPLICATE_SLUG", "A product with the same slug already exists.");
        }

        if (!string.IsNullOrWhiteSpace(request.BrandId) && !await uow.Brands.ExistsAsync(request.BrandId, cancellationToken))
            throw new DomainException("BRAND_NOT_FOUND", $"Brand '{request.BrandId}' does not exist.");

        foreach (var categoryId in request.CategoryIds)
        {
            if (await uow.Categories.GetByIdAsync(categoryId, cancellationToken) is null)
                throw new DomainException("CATEGORY_NOT_FOUND", $"Category '{categoryId}' does not exist.");
        }

        var product = new Product
        {
            Sku = request.Sku.Trim(),
            Slug = request.Slug.Trim().ToLowerInvariant(),
            Title = request.Title.Trim(),
            Description = request.Description?.Trim(),
            BrandId = request.BrandId?.Trim(),
            CategoryIds = request.CategoryIds,
            Price = new Money { Amount = request.Amount, Currency = request.Currency.Trim().ToUpperInvariant() },
        };

        await uow.Products.SaveAsync(product, isNew: true, cancellationToken);

        var variantData = product.Variants.Select(v => new ProductVariantEventData(v.VariantId, v.Sku, v.Attributes)).ToList();
        var domainEvent = new ProductCreatedEvent(product.Id, product.Sku, product.Title, product.BrandId, product.CategoryIds, product.Price.Amount, product.Price.Currency, variantData);
        await eventPublisher.PublishAsync(domainEvent, cancellationToken);

        logger.LogInformation("Product {ProductId} created with SKU {Sku} in {ElapsedMs}ms", product.Id, product.Sku, sw.ElapsedMilliseconds);
        return ProductResponseDto.FromEntity(product);
    }
}

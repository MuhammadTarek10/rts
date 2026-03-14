using Catalog.Api.Application.DTOs;
using Catalog.Api.Domain.Entities;
using Catalog.Api.Domain.Events;
using Catalog.Api.Infrastructure;
using Catalog.Api.Infrastructure.Messaging;
using Catalog.Api.Infrastructure.Services;
using Catalog.Api.Shared.Exceptions;

namespace Catalog.Api.Application.Handlers;

public sealed class UpdateProductHandler(
    CatalogUnitOfWork uow,
    ICacheService cache,
    ICatalogEventPublisher eventPublisher,
    ILogger<UpdateProductHandler> logger)
{
    public async Task<ProductResponseDto> HandleAsync(
        string id,
        UpdateProductDto request,
        CancellationToken cancellationToken = default)
    {
        var sw = System.Diagnostics.Stopwatch.StartNew();
        logger.LogInformation("Updating product {ProductId}", id);

        var product = await uow.Products.GetByIdAsync(id, cancellationToken)
            ?? throw new NotFoundException("PRODUCT_NOT_FOUND", $"Product '{id}' not found.");

        if (!string.IsNullOrWhiteSpace(request.BrandId) && !await uow.Brands.ExistsAsync(request.BrandId, cancellationToken))
            throw new DomainException("BRAND_NOT_FOUND", $"Brand '{request.BrandId}' does not exist.");

        foreach (var categoryId in request.CategoryIds)
        {
            if (await uow.Categories.GetByIdAsync(categoryId, cancellationToken) is null)
                throw new DomainException("CATEGORY_NOT_FOUND", $"Category '{categoryId}' does not exist.");
        }

        var changedFields = new List<string>();
        if (product.Title != request.Title) changedFields.Add("Title");
        if (product.Description != request.Description) changedFields.Add("Description");
        if (product.BrandId != request.BrandId) changedFields.Add("BrandId");
        if (!product.CategoryIds.SequenceEqual(request.CategoryIds)) changedFields.Add("CategoryIds");
        if (product.Price.Amount != request.Amount || product.Price.Currency != request.Currency) changedFields.Add("Price");
        if (!product.Tags.SequenceEqual(request.Tags)) changedFields.Add("Tags");

        product.Title = request.Title.Trim();
        product.Description = request.Description?.Trim();
        product.BrandId = request.BrandId?.Trim();
        product.CategoryIds = request.CategoryIds;
        product.Price = new Money { Amount = request.Amount, Currency = request.Currency.Trim().ToUpperInvariant() };
        product.Tags = request.Tags;
        product.Touch();

        await uow.Products.ReplaceIfVersionMatchesAsync(product, request.Version, cancellationToken);

        await cache.RemoveAsync($"catalog:product:{id}", cancellationToken);
        await cache.RemoveAsync($"catalog:product:slug:{product.Slug}", cancellationToken);

        if (changedFields.Count > 0)
            await eventPublisher.PublishAsync(new ProductUpdatedEvent(product.Id, changedFields), cancellationToken);

        logger.LogInformation("Product {ProductId} updated in {ElapsedMs}ms", id, sw.ElapsedMilliseconds);
        return ProductResponseDto.FromEntity(product);
    }
}

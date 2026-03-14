using Catalog.Api.Domain.Events;
using Catalog.Api.Infrastructure;
using Catalog.Api.Infrastructure.Messaging;
using Catalog.Api.Infrastructure.Services;
using Catalog.Api.Shared.Exceptions;

namespace Catalog.Api.Application.Handlers;

public sealed class DeleteProductHandler(
    CatalogUnitOfWork uow,
    ICacheService cache,
    ICatalogEventPublisher eventPublisher,
    ILogger<DeleteProductHandler> logger)
{
    public async Task HandleAsync(string id, CancellationToken cancellationToken = default)
    {
        logger.LogInformation("Deleting (archiving) product {ProductId}", id);

        var product = await uow.Products.GetByIdAsync(id, cancellationToken)
            ?? throw new NotFoundException("PRODUCT_NOT_FOUND", $"Product '{id}' not found.");

        var oldStatus = product.Status;
        product.SetStatus(Domain.Entities.ProductStatus.Archived);
        await uow.Products.UpdateAsync(product, cancellationToken);

        await cache.RemoveAsync($"catalog:product:{id}", cancellationToken);
        await cache.RemoveAsync($"catalog:product:slug:{product.Slug}", cancellationToken);

        await eventPublisher.PublishAsync(new ProductDeletedEvent(product.Id, product.Sku), cancellationToken);

        logger.LogInformation("Product {ProductId} archived (was {OldStatus})", id, oldStatus);
    }
}

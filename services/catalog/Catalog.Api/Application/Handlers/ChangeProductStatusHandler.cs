using Catalog.Api.Application.DTOs;
using Catalog.Api.Domain.Entities;
using Catalog.Api.Domain.Events;
using Catalog.Api.Infrastructure;
using Catalog.Api.Infrastructure.Messaging;
using Catalog.Api.Infrastructure.Services;
using Catalog.Api.Shared.Exceptions;

namespace Catalog.Api.Application.Handlers;

public sealed class ChangeProductStatusHandler(
    CatalogUnitOfWork uow,
    ICacheService cache,
    ICatalogEventPublisher eventPublisher,
    ILogger<ChangeProductStatusHandler> logger)
{
    public async Task<ProductResponseDto> HandleAsync(
        string id,
        ChangeStatusDto request,
        CancellationToken cancellationToken = default)
    {
        logger.LogInformation("Changing status of product {ProductId} to {NewStatus}", id, request.Status);

        if (!Enum.TryParse<ProductStatus>(request.Status, ignoreCase: true, out var newStatus))
            throw new DomainException("INVALID_STATUS", $"'{request.Status}' is not a valid product status.");

        var product = await uow.Products.GetByIdAsync(id, cancellationToken)
            ?? throw new NotFoundException("PRODUCT_NOT_FOUND", $"Product '{id}' not found.");

        var oldStatus = product.Status;
        product.SetStatus(newStatus);
        await uow.Products.UpdateAsync(product, cancellationToken);

        await cache.RemoveAsync($"catalog:product:{id}", cancellationToken);
        await cache.RemoveAsync($"catalog:product:slug:{product.Slug}", cancellationToken);

        await eventPublisher.PublishAsync(new ProductStatusChangedEvent(product.Id, oldStatus.ToString(), newStatus.ToString()), cancellationToken);

        logger.LogInformation("Product {ProductId} status changed from {OldStatus} to {NewStatus}", id, oldStatus, newStatus);
        return ProductResponseDto.FromEntity(product);
    }
}

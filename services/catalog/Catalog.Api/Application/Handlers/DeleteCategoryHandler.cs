using Catalog.Api.Infrastructure;
using Catalog.Api.Infrastructure.Services;
using Catalog.Api.Shared.Exceptions;

namespace Catalog.Api.Application.Handlers;

public sealed class DeleteCategoryHandler(
    CatalogUnitOfWork uow,
    ICacheService cache,
    ILogger<DeleteCategoryHandler> logger)
{
    public async Task HandleAsync(string id, CancellationToken cancellationToken = default)
    {
        logger.LogInformation("Deleting category {CategoryId}", id);

        var category = await uow.Categories.GetByIdAsync(id, cancellationToken)
            ?? throw new NotFoundException("CATEGORY_NOT_FOUND", $"Category '{id}' not found.");

        if (await uow.Categories.HasProductsAsync(id, cancellationToken))
            throw new ConflictException("CATEGORY_HAS_PRODUCTS", "Cannot delete category: products reference it.");

        if (await uow.Categories.HasChildrenAsync(id, cancellationToken))
            throw new ConflictException("CATEGORY_HAS_CHILDREN", "Cannot delete category: it has child categories.");

        await uow.Categories.DeleteAsync(id, cancellationToken);
        await cache.RemoveAsync("catalog:categories:tree", cancellationToken);

        logger.LogInformation("Category {CategoryId} deleted", id);
    }
}

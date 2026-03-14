using Catalog.Api.Application.DTOs;
using Catalog.Api.Infrastructure;
using Catalog.Api.Infrastructure.Services;
using Catalog.Api.Shared.Exceptions;

namespace Catalog.Api.Application.Handlers;

public sealed class UpdateCategoryHandler(
    CatalogUnitOfWork uow,
    ICacheService cache,
    ILogger<UpdateCategoryHandler> logger)
{
    public async Task<CategoryResponseDto> HandleAsync(
        string id,
        UpdateCategoryDto request,
        CancellationToken cancellationToken = default)
    {
        logger.LogInformation("Updating category {CategoryId}", id);

        var category = await uow.Categories.GetByIdAsync(id, cancellationToken)
            ?? throw new NotFoundException("CATEGORY_NOT_FOUND", $"Category '{id}' not found.");

        // Validate no circular reference if parent is changing
        if (request.ParentId != category.ParentId && !string.IsNullOrWhiteSpace(request.ParentId))
        {
            var newParent = await uow.Categories.GetByIdAsync(request.ParentId, cancellationToken)
                ?? throw new NotFoundException("PARENT_NOT_FOUND", $"Parent category '{request.ParentId}' not found.");

            // Check if current category appears in new parent's path (circular)
            if (newParent.Path.Contains(id) || newParent.Id == id)
                throw new DomainException("CIRCULAR_REFERENCE", "Cannot set parent: would create a circular reference.");

            category.Path = [..newParent.Path, newParent.Id];
            category.Depth = newParent.Depth + 1;
        }
        else if (string.IsNullOrWhiteSpace(request.ParentId))
        {
            category.Path = [];
            category.Depth = 0;
        }

        category.Name = request.Name.Trim();
        category.Description = request.Description?.Trim();
        category.ParentId = string.IsNullOrWhiteSpace(request.ParentId) ? null : request.ParentId.Trim();
        category.SortOrder = request.SortOrder;
        category.IsActive = request.IsActive;
        category.ImageUrl = request.ImageUrl;
        category.UpdatedAt = DateTime.UtcNow;

        await uow.Categories.UpdateAsync(category, cancellationToken);
        await cache.RemoveAsync("catalog:categories:tree", cancellationToken);

        logger.LogInformation("Category {CategoryId} updated", id);
        return CategoryResponseDto.FromEntity(category);
    }
}

using Catalog.Api.Application.DTOs;
using Catalog.Api.Domain.Entities;
using Catalog.Api.Domain.Events;
using Catalog.Api.Infrastructure;
using Catalog.Api.Infrastructure.Messaging;
using Catalog.Api.Infrastructure.Services;
using Catalog.Api.Shared.Exceptions;

namespace Catalog.Api.Application.Handlers;

public sealed class CreateCategoryHandler(
    CatalogUnitOfWork uow,
    ICacheService cache,
    ICatalogEventPublisher eventPublisher,
    ILogger<CreateCategoryHandler> logger)
{
    public async Task<CategoryResponseDto> HandleAsync(
        CreateCategoryDto request,
        CancellationToken cancellationToken = default)
    {
        logger.LogInformation("Creating category {Name}", request.Name);

        if (await uow.Categories.ExistsBySlugAsync(request.Slug, cancellationToken))
            throw new ConflictException("DUPLICATE_SLUG", "A category with this slug already exists.");

        IReadOnlyList<string> path = [];
        int depth = 0;

        if (!string.IsNullOrWhiteSpace(request.ParentId))
        {
            var parent = await uow.Categories.GetByIdAsync(request.ParentId, cancellationToken)
                ?? throw new NotFoundException("PARENT_NOT_FOUND", $"Parent category '{request.ParentId}' not found.");
            // Path = parent's path + parent's id
            path = [..parent.Path, parent.Id];
            depth = parent.Depth + 1;
        }

        var category = new Category
        {
            Name = request.Name.Trim(),
            Slug = request.Slug.Trim().ToLowerInvariant(),
            Description = request.Description?.Trim(),
            ParentId = request.ParentId?.Trim(),
            Path = path,
            Depth = depth,
            SortOrder = request.SortOrder,
            ImageUrl = request.ImageUrl,
        };

        await uow.Categories.SaveAsync(category, isNew: true, cancellationToken);

        await cache.RemoveAsync("catalog:categories:tree", cancellationToken);

        await eventPublisher.PublishAsync(new CategoryCreatedEvent(category.Id, category.Name, category.ParentId), cancellationToken);

        logger.LogInformation("Category {CategoryId} created: {Name}", category.Id, category.Name);
        return CategoryResponseDto.FromEntity(category);
    }
}

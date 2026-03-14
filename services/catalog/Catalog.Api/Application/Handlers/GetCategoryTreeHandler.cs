using Catalog.Api.Application.DTOs;
using Catalog.Api.Infrastructure;
using Catalog.Api.Infrastructure.Services;

namespace Catalog.Api.Application.Handlers;

public sealed class GetCategoryTreeHandler(
    CatalogUnitOfWork uow,
    ICacheService cache,
    ILogger<GetCategoryTreeHandler> logger)
{
    public async Task<IReadOnlyList<CategoryTreeDto>> HandleAsync(CancellationToken cancellationToken = default)
    {
        const string cacheKey = "catalog:categories:tree";
        var cached = await cache.GetAsync<List<CategoryTreeDto>>(cacheKey, cancellationToken);
        if (cached is not null) return cached;

        var allCategories = await uow.Categories.GetAllActiveAsync(cancellationToken);
        var tree = BuildTree(allCategories, null);

        await cache.SetAsync(cacheKey, tree, TimeSpan.FromMinutes(15), cancellationToken);

        logger.LogInformation("Category tree built with {Count} root categories", tree.Count);
        return tree;
    }

    private static List<CategoryTreeDto> BuildTree(IReadOnlyList<Domain.Entities.Category> categories, string? parentId)
    {
        return categories
            .Where(c => c.ParentId == parentId)
            .OrderBy(c => c.SortOrder)
            .Select(c => new CategoryTreeDto(
                c.Id, c.Name, c.Slug, c.Description, c.ParentId,
                c.SortOrder, c.IsActive, c.ImageUrl,
                BuildTree(categories, c.Id)))
            .ToList();
    }
}

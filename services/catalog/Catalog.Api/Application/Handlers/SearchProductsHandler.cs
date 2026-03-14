using System.Security.Cryptography;
using System.Text;
using System.Text.Json;
using Catalog.Api.Application.DTOs;
using Catalog.Api.Domain.Models;
using Catalog.Api.Infrastructure;
using Catalog.Api.Infrastructure.Services;

namespace Catalog.Api.Application.Handlers;

public sealed class SearchProductsHandler(
    CatalogUnitOfWork uow,
    ICacheService cache,
    ILogger<SearchProductsHandler> logger)
{
    private const int MaxPageSize = 100;

    public async Task<SearchProductsResponse> HandleAsync(
        SearchProductsRequest request,
        CancellationToken cancellationToken = default)
    {
        var pageSize = Math.Min(request.PageSize, MaxPageSize);
        var page = Math.Max(request.Page, 1);

        // Expand category to include descendants
        IReadOnlyList<string>? categoryIds = null;
        if (!string.IsNullOrWhiteSpace(request.CategoryId))
        {
            var descendants = await uow.Categories.GetDescendantIdsAsync(request.CategoryId, cancellationToken);
            categoryIds = new[] { request.CategoryId }.Concat(descendants).ToList();
        }

        // Cache search results (TTL-only, no active invalidation)
        var cacheKey = BuildCacheKey(request, pageSize, page, categoryIds);
        var cached = await cache.GetAsync<SearchProductsResponse>(cacheKey, cancellationToken);
        if (cached is not null) return cached;

        var criteria = new ProductSearchCriteria
        {
            Query = request.Query,
            CategoryId = request.CategoryId,
            CategoryIds = categoryIds,
            BrandId = request.BrandId,
            MinPrice = request.MinPrice,
            MaxPrice = request.MaxPrice,
            Status = !string.IsNullOrWhiteSpace(request.Status)
                ? Enum.TryParse<Domain.Entities.ProductStatus>(request.Status, true, out var s) ? s : null
                : null,
            SortBy = request.SortBy,
            Page = page,
            PageSize = pageSize,
        };

        var (items, totalCount) = await uow.Products.SearchAsync(criteria, cancellationToken);
        var totalPages = (int)Math.Ceiling((double)totalCount / pageSize);

        var response = new SearchProductsResponse(
            items.Select(ProductResponseDto.FromEntity).ToList(),
            totalCount,
            page,
            pageSize,
            totalPages
        );

        await cache.SetAsync(cacheKey, response, TimeSpan.FromMinutes(2), cancellationToken);

        logger.LogInformation("Search returned {Count}/{Total} products", items.Count, totalCount);
        return response;
    }

    private static string BuildCacheKey(SearchProductsRequest req, int pageSize, int page, IReadOnlyList<string>? categoryIds)
    {
        var keyData = JsonSerializer.Serialize(new { req.Query, req.CategoryId, categoryIds, req.BrandId, req.MinPrice, req.MaxPrice, req.Status, req.SortBy, page, pageSize });
        var hash = Convert.ToHexString(SHA256.HashData(Encoding.UTF8.GetBytes(keyData)))[..16];
        return $"catalog:search:{hash}";
    }
}

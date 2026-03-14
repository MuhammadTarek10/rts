using Catalog.Api.Domain.Entities;

namespace Catalog.Api.Domain.Models;

public sealed class ProductSearchCriteria
{
    public string? Query { get; init; }
    public string? CategoryId { get; init; }
    public IReadOnlyList<string>? CategoryIds { get; init; } // expanded descendants
    public string? BrandId { get; init; }
    public decimal? MinPrice { get; init; }
    public decimal? MaxPrice { get; init; }
    public ProductStatus? Status { get; init; }
    public string? SortBy { get; init; } // price_asc, price_desc, newest, relevance
    public int Page { get; init; } = 1;
    public int PageSize { get; init; } = 20;
}

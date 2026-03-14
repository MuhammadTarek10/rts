namespace Catalog.Api.Application.DTOs;

public sealed class SearchProductsRequest
{
    public string? Query { get; init; }
    public string? CategoryId { get; init; }
    public string? BrandId { get; init; }
    public decimal? MinPrice { get; init; }
    public decimal? MaxPrice { get; init; }
    public string? Status { get; init; }
    public string? SortBy { get; init; }
    public int Page { get; init; } = 1;
    public int PageSize { get; init; } = 20;
}

public sealed record SearchProductsResponse(
    IReadOnlyList<ProductResponseDto> Items,
    long TotalCount,
    int Page,
    int PageSize,
    int TotalPages
);

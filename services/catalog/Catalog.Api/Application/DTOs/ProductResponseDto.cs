using Catalog.Api.Domain.Entities;

namespace Catalog.Api.Application.DTOs;

public sealed record ProductResponseDto(
    string Id,
    string Sku,
    string Slug,
    string Title,
    string? Description,
    decimal Amount,
    string Currency,
    string Status,
    string? BrandId,
    IReadOnlyList<string> CategoryIds,
    IReadOnlyList<ProductImageDto> Images,
    decimal? AverageRating,
    int ReviewCount,
    IReadOnlyList<string> Tags,
    DateTime CreatedAt,
    DateTime UpdatedAt,
    int Version
)
{
    public static ProductResponseDto FromEntity(Product product) => new(
        product.Id,
        product.Sku,
        product.Slug,
        product.Title,
        product.Description,
        product.Price.Amount,
        product.Price.Currency,
        product.Status.ToString(),
        product.BrandId,
        product.CategoryIds,
        product.Images.Select(ProductImageDto.FromEntity).ToList(),
        product.AverageRating,
        product.ReviewCount,
        product.Tags,
        product.CreatedAt,
        product.UpdatedAt,
        product.Version
    );
}

public sealed record ProductImageDto(
    string ImageId,
    string Url,
    string? AltText,
    int SortOrder,
    bool IsPrimary
)
{
    public static ProductImageDto FromEntity(ProductImage image) => new(
        image.ImageId,
        image.Url,
        image.AltText,
        image.SortOrder,
        image.IsPrimary
    );
}

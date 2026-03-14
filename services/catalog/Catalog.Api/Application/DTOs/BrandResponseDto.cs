using Catalog.Api.Domain.Entities;

namespace Catalog.Api.Application.DTOs;

public sealed record BrandResponseDto(
    string Id,
    string Name,
    string Slug,
    string? Description,
    string? LogoUrl,
    string? Website,
    bool IsActive,
    DateTime CreatedAt,
    DateTime UpdatedAt
)
{
    public static BrandResponseDto FromEntity(Brand brand) => new(
        brand.Id,
        brand.Name,
        brand.Slug,
        brand.Description,
        brand.LogoUrl,
        brand.Website,
        brand.IsActive,
        brand.CreatedAt,
        brand.UpdatedAt
    );
}

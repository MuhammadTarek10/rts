using Catalog.Api.Domain.Entities;

namespace Catalog.Api.Application.DTOs;

public sealed record CategoryResponseDto(
    string Id,
    string Name,
    string Slug,
    string? Description,
    string? ParentId,
    IReadOnlyList<string> Path,
    int Depth,
    int SortOrder,
    bool IsActive,
    string? ImageUrl,
    DateTime CreatedAt,
    DateTime UpdatedAt
)
{
    public static CategoryResponseDto FromEntity(Category category) => new(
        category.Id,
        category.Name,
        category.Slug,
        category.Description,
        category.ParentId,
        category.Path,
        category.Depth,
        category.SortOrder,
        category.IsActive,
        category.ImageUrl,
        category.CreatedAt,
        category.UpdatedAt
    );
}

public sealed record CategoryTreeDto(
    string Id,
    string Name,
    string Slug,
    string? Description,
    string? ParentId,
    int SortOrder,
    bool IsActive,
    string? ImageUrl,
    IReadOnlyList<CategoryTreeDto> Children
);

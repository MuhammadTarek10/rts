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
    DateTime CreatedAt,
    DateTime UpdatedAt,
    int Version
);

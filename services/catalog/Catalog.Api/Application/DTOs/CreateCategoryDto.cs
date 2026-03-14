using System.ComponentModel.DataAnnotations;

namespace Catalog.Api.Application.DTOs;

public sealed class CreateCategoryDto
{
    [Required][MaxLength(100)] public string Name { get; init; } = string.Empty;
    [Required][MaxLength(120)] public string Slug { get; init; } = string.Empty;
    [MaxLength(500)] public string? Description { get; init; }
    public string? ParentId { get; init; }
    public int SortOrder { get; init; }
    public string? ImageUrl { get; init; }
}

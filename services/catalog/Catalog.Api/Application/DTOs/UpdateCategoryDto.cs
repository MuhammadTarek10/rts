using System.ComponentModel.DataAnnotations;

namespace Catalog.Api.Application.DTOs;

public sealed class UpdateCategoryDto
{
    [Required][MaxLength(100)] public string Name { get; init; } = string.Empty;
    [MaxLength(500)] public string? Description { get; init; }
    public string? ParentId { get; init; }
    public int SortOrder { get; init; }
    public bool IsActive { get; init; } = true;
    public string? ImageUrl { get; init; }
}

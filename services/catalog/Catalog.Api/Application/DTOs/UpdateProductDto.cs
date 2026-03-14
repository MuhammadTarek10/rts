using System.ComponentModel.DataAnnotations;

namespace Catalog.Api.Application.DTOs;

public sealed class UpdateProductDto
{
    [Required][MaxLength(200)] public string Title { get; init; } = string.Empty;
    [MaxLength(2000)] public string? Description { get; init; }
    public string? BrandId { get; init; }
    public IReadOnlyList<string> CategoryIds { get; init; } = [];
    [Range(0.01, double.MaxValue)] public decimal Amount { get; init; }
    [Required][Length(3, 3)] public string Currency { get; init; } = "USD";
    public IReadOnlyList<string> Tags { get; init; } = [];
    public int Version { get; init; } // for optimistic concurrency
}

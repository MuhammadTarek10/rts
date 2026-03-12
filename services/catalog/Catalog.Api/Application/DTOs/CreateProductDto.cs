using System.ComponentModel.DataAnnotations;

namespace Catalog.Api.Application.DTOs;

public sealed class CreateProductDto
{
	[Required]
	[MaxLength(100)]
	public string Sku { get; init; } = string.Empty;

	[Required]
	[MaxLength(120)]
	public string Slug { get; init; } = string.Empty;

	[Required]
	[MaxLength(200)]
	public string Title { get; init; } = string.Empty;

	[MaxLength(2000)]
	public string? Description { get; init; }

	[Range(0.01, double.MaxValue)]
	public decimal Amount { get; init; }

	[Required]
	[Length(3, 3)]
	public string Currency { get; init; } = "USD";

	public string? BrandId { get; init; }
	public IReadOnlyList<string> CategoryIds { get; init; } = [];
}

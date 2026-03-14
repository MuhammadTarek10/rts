using Catalog.Api.Shared.Abstraction;
using MongoDB.Bson;
using MongoDB.Bson.Serialization.Attributes;

namespace Catalog.Api.Domain.Entities;

/// <summary>
/// Represents a catalog product with pricing, variants, and lifecycle status.
/// </summary>
public sealed class Product : IEntity
{
    [BsonId]
    [BsonRepresentation(BsonType.String)]
    public string Id { get; init; } = Guid.CreateVersion7().ToString();

    [BsonElement("sku")]
    public required string Sku { get; init; }

    [BsonElement("slug")]
    public required string Slug { get; init; }

    [BsonElement("title")]
    public required string Title { get; set; }

    [BsonElement("description")]
    public string? Description { get; set; }

    [BsonElement("brandId")]
    public string? BrandId { get; set; }

    [BsonElement("categoryIds")]
    public IReadOnlyList<string> CategoryIds { get; set; } = [];

    [BsonElement("status")]
    [BsonRepresentation(BsonType.String)]
    public ProductStatus Status { get; private set; } = ProductStatus.Draft;

    [BsonElement("price")]
    public required Money Price { get; set; }

    [BsonElement("variants")]
    public IReadOnlyList<ProductVariant> Variants { get; set; } = [];

    [BsonElement("images")]
    public IReadOnlyList<ProductImage> Images { get; set; } = [];

    [BsonElement("averageRating")]
    public decimal? AverageRating { get; set; }

    [BsonElement("reviewCount")]
    public int ReviewCount { get; set; } = 0;

    [BsonElement("tags")]
    public IReadOnlyList<string> Tags { get; set; } = [];

    [BsonElement("createdAt")]
    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

    [BsonElement("updatedAt")]
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;

    [BsonElement("version")]
    public int Version { get; private set; } = 1;

    /// <summary>
    /// Transitions the product to the specified lifecycle status.
    /// </summary>
    public void SetStatus(ProductStatus status)
    {
        Status = status;
        Touch();
    }

    /// <summary>
    /// Updates the modification timestamp and increments the version.
    /// </summary>
    public void Touch()
    {
        UpdatedAt = DateTime.UtcNow;
        Version++;
    }
}

public sealed class ProductVariant
{
    [BsonElement("variantId")]
    public string VariantId { get; init; } = Guid.CreateVersion7().ToString();

    [BsonElement("sku")]
    public required string Sku { get; init; }

    [BsonElement("attributes")]
    public Dictionary<string, string> Attributes { get; init; } = [];

    [BsonElement("price")]
    public Money? Price { get; init; }
}

public sealed class ProductImage
{
    [BsonElement("imageId")]
    public string ImageId { get; init; } = Guid.CreateVersion7().ToString();

    [BsonElement("url")]
    public required string Url { get; init; }

    [BsonElement("key")]
    public required string Key { get; init; }

    [BsonElement("altText")]
    public string? AltText { get; init; }

    [BsonElement("sortOrder")]
    public int SortOrder { get; init; }

    [BsonElement("isPrimary")]
    public bool IsPrimary { get; init; }
}

public sealed class Money
{
    [BsonElement("amount")]
    [BsonRepresentation(BsonType.Decimal128)]
    public decimal Amount { get; init; }

    [BsonElement("currency")]
    public string Currency { get; init; } = "USD";
}

public enum ProductStatus
{
    Draft,
    Active,
    Archived,
}

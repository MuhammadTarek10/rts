using Catalog.Api.Shared.Abstraction;
using MongoDB.Bson;
using MongoDB.Bson.Serialization.Attributes;

namespace Catalog.Api.Domain.Entities;

public sealed class Category : IEntity
{
    [BsonId]
    [BsonRepresentation(BsonType.String)]
    public string Id { get; init; } = Guid.CreateVersion7().ToString();

    [BsonElement("name")]
    public required string Name { get; set; }

    [BsonElement("slug")]
    public required string Slug { get; init; }

    [BsonElement("description")]
    public string? Description { get; set; }

    [BsonElement("parentId")]
    public string? ParentId { get; set; }

    // Materialized path: list of ancestor IDs from root to this category's parent
    [BsonElement("path")]
    public IReadOnlyList<string> Path { get; set; } = [];

    [BsonElement("depth")]
    public int Depth { get; set; }

    [BsonElement("sortOrder")]
    public int SortOrder { get; set; }

    [BsonElement("isActive")]
    public bool IsActive { get; set; } = true;

    [BsonElement("imageUrl")]
    public string? ImageUrl { get; set; }

    [BsonElement("createdAt")]
    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

    [BsonElement("updatedAt")]
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
}

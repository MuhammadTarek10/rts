using Catalog.Api.Shared.Abstraction;
using MongoDB.Bson;
using MongoDB.Bson.Serialization.Attributes;

namespace Catalog.Api.Domain.Entities;

public sealed class Brand : IEntity
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

    [BsonElement("logoUrl")]
    public string? LogoUrl { get; set; }

    [BsonElement("website")]
    public string? Website { get; set; }

    [BsonElement("isActive")]
    public bool IsActive { get; set; } = true;

    [BsonElement("createdAt")]
    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

    [BsonElement("updatedAt")]
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
}

namespace Catalog.Api.Application.DTOs;

public sealed class ReorderImagesDto
{
    public IReadOnlyList<ImageOrderItem> Order { get; init; } = [];
}

public sealed class ImageOrderItem
{
    public string ImageId { get; init; } = string.Empty;
    public int SortOrder { get; init; }
}

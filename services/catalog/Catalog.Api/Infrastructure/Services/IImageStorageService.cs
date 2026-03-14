namespace Catalog.Api.Infrastructure.Services;

public interface IImageStorageService
{
    Task<(string Url, string Key)> UploadAsync(
        Stream stream,
        string fileName,
        string contentType,
        string productId,
        CancellationToken cancellationToken = default);

    Task DeleteAsync(string key, CancellationToken cancellationToken = default);
}

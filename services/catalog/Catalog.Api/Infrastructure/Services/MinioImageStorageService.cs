using Amazon.S3;
using Amazon.S3.Model;
using Catalog.Api.Infrastructure.Settings;
using Microsoft.Extensions.Options;

namespace Catalog.Api.Infrastructure.Services;

public sealed class MinioImageStorageService(
    IAmazonS3 s3Client,
    IOptions<MinioSettings> options,
    ILogger<MinioImageStorageService> logger) : IImageStorageService
{
    private readonly MinioSettings _settings = options.Value;

    public async Task<(string Url, string Key)> UploadAsync(
        Stream stream,
        string fileName,
        string contentType,
        string productId,
        CancellationToken cancellationToken = default)
    {
        var extension = Path.GetExtension(fileName).TrimStart('.').ToLowerInvariant();
        var imageId = Guid.CreateVersion7().ToString();
        var key = $"products/{productId}/{imageId}.{extension}";

        var request = new PutObjectRequest
        {
            BucketName = _settings.BucketName,
            Key = key,
            InputStream = stream,
            ContentType = contentType,
            AutoCloseStream = false,
        };

        await s3Client.PutObjectAsync(request, cancellationToken);
        logger.LogInformation("Image uploaded to MinIO: {Key}", key);

        var url = $"{_settings.PublicBaseUrl.TrimEnd('/')}/{key}";
        return (url, key);
    }

    public async Task DeleteAsync(string key, CancellationToken cancellationToken = default)
    {
        var request = new DeleteObjectRequest
        {
            BucketName = _settings.BucketName,
            Key = key,
        };

        await s3Client.DeleteObjectAsync(request, cancellationToken);
        logger.LogInformation("Image deleted from MinIO: {Key}", key);
    }
}

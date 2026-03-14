namespace Catalog.Api.Infrastructure.Settings;

public sealed class MinioSettings
{
    public string Endpoint { get; set; } = string.Empty;
    public string AccessKey { get; set; } = string.Empty;
    public string SecretKey { get; set; } = string.Empty;
    public string BucketName { get; set; } = "catalog-images";
    public bool UseSSL { get; set; }
    public string PublicBaseUrl { get; set; } = string.Empty;
}

namespace Catalog.Api.Infrastructure.Settings;

/// <summary>
/// Strongly-typed configuration for the catalog MongoDB connection.
/// </summary>
public sealed class CatalogMongoSettings
{
    /// <summary>
    /// MongoDB connection string.
    /// </summary>
    public string ConnectionString { get; set; } = string.Empty;

    /// <summary>
    /// Name of the MongoDB database to use.
    /// </summary>
    public string DatabaseName { get; set; } = "catalog";
}

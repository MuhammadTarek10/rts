using Catalog.Api.Application.Handlers;
using Catalog.Api.Domain.Interfaces;
using Catalog.Api.Infrastructure.Repositories;
using Catalog.Api.Infrastructure.Settings;
using MongoDB.Driver;

namespace Catalog.Api.Infrastructure.Extensions;

public static class ServiceExtensions
{
    public static void AddInfrastructure(this IServiceCollection services, IConfiguration configuration)
    {
        services.Configure<CatalogMongoSettings>(settings =>
        {
            var section = configuration.GetSection("Catalog:Mongo");

            settings.ConnectionString = section["ConnectionString"]
                ?? configuration.GetConnectionString("CatalogMongo")
                ?? string.Empty;

            settings.DatabaseName = section["DatabaseName"]
                ?? configuration["CatalogMongoDatabase"]
                ?? configuration["ConnectionStrings:DatabaseName"]
                ?? "catalog";
        });

        var connectionString = configuration.GetSection("Catalog:Mongo")["ConnectionString"]
            ?? configuration.GetConnectionString("CatalogMongo")!;

        var databaseName = configuration.GetSection("Catalog:Mongo")["DatabaseName"]
            ?? configuration["CatalogMongoDatabase"]
            ?? configuration["ConnectionStrings:DatabaseName"]
            ?? "catalog";

        services.AddSingleton<IMongoClient>(_ => new MongoClient(connectionString));
        services.AddSingleton<IMongoDatabase>(serviceProvider =>
        {
            var client = serviceProvider.GetRequiredService<IMongoClient>();
            return client.GetDatabase(databaseName);
        });

        services.AddScoped<IProductRepository, ProductRepository>();
        services.AddScoped<CreateProductHandler>();
        services.AddHostedService<MongoIndexesInitializer>();
    }
}

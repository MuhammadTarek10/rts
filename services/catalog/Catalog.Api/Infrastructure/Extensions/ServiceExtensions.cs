using Catalog.Api.Infrastructure;
using Microsoft.EntityFrameworkCore;

namespace Catalog.Api.Infrastructure.Extensions;

public static class ServiceExtensions
{

    public static void AddInfrastructure(this IServiceCollection services, IConfiguration configuration)
    {
        var connectionString = configuration.GetConnectionString("CatalogMongo")!;
        var databaseName =
            configuration["CatalogMongoDatabase"]
            ?? configuration["ConnectionStrings:DatabaseName"]
            ?? "catalog";

        services.AddDbContext<AppDbContext>(options =>
            options.UseMongoDB(connectionString, databaseName));
    }
}
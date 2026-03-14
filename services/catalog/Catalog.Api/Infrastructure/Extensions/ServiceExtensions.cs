using Amazon.S3;
using Catalog.Api.Application.Handlers;
using RabbitMQ.Client;
using Catalog.Api.Domain.Interfaces;
using Catalog.Api.Infrastructure.Messaging;
using Catalog.Api.Infrastructure.Repositories;
using Catalog.Api.Infrastructure.Services;
using Catalog.Api.Infrastructure.Settings;
using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.IdentityModel.Tokens;
using MongoDB.Driver;
using System.Text;

namespace Catalog.Api.Infrastructure.Extensions;

public static class ServiceExtensions
{
    public static void AddInfrastructure(this IServiceCollection services, IConfiguration configuration)
    {
        // MongoDB
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
        services.AddSingleton<IMongoDatabase>(sp =>
            sp.GetRequiredService<IMongoClient>().GetDatabase(databaseName));

        // Repositories & UnitOfWork
        services.AddScoped<IProductRepository, ProductRepository>();
        services.AddScoped<ICategoryRepository, CategoryRepository>();
        services.AddScoped<IBrandRepository, BrandRepository>();
        services.AddScoped<CatalogUnitOfWork>();
        services.AddHostedService<MongoIndexesInitializer>();

        // JWT Settings
        services.Configure<JwtSettings>(settings =>
        {
            settings.AccessSecret = configuration["Jwt:AccessSecret"]
                ?? configuration["JWT_ACCESS_SECRET"]
                ?? string.Empty;
        });

        var jwtSecret = configuration["Jwt:AccessSecret"]
            ?? configuration["JWT_ACCESS_SECRET"]
            ?? string.Empty;

        services
            .AddAuthentication(JwtBearerDefaults.AuthenticationScheme)
            .AddJwtBearer(options =>
            {
                options.TokenValidationParameters = new TokenValidationParameters
                {
                    ValidateIssuerSigningKey = true,
                    IssuerSigningKey = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(jwtSecret)),
                    ValidateIssuer = false,
                    ValidateAudience = false,
                    ValidateLifetime = true,
                    ClockSkew = TimeSpan.FromSeconds(30),
                };
                options.Events = new JwtBearerEvents
                {
                    OnMessageReceived = ctx =>
                    {
                        if (ctx.Request.Cookies.TryGetValue("access_token", out var cookieToken))
                            ctx.Token = cookieToken;
                        return Task.CompletedTask;
                    },
                    OnChallenge = async ctx =>
                    {
                        ctx.HandleResponse();
                        ctx.Response.StatusCode = 401;
                        ctx.Response.ContentType = "application/json";
                        await ctx.Response.WriteAsync("""{"code":"UNAUTHORIZED","message":"Authentication is required."}""");
                    },
                    OnForbidden = async ctx =>
                    {
                        ctx.Response.StatusCode = 403;
                        ctx.Response.ContentType = "application/json";
                        await ctx.Response.WriteAsync("""{"code":"FORBIDDEN","message":"You do not have permission to access this resource."}""");
                    },
                };
            });

        // Authorization policies
        services.AddAuthorizationBuilder()
            .AddPolicy("Admin", policy => policy.RequireClaim("role", "admin"));

        // Redis Cache
        var redisConnectionString = configuration["Redis:ConnectionString"] ?? "localhost:6379";
        services.AddStackExchangeRedisCache(opts => opts.Configuration = redisConnectionString);
        services.AddScoped<ICacheService, RedisCacheService>();

        // MinIO (S3-compatible)
        services.Configure<MinioSettings>(configuration.GetSection("Minio"));
        var minioSettings = configuration.GetSection("Minio").Get<MinioSettings>() ?? new MinioSettings();
        services.AddSingleton<IAmazonS3>(_ => new AmazonS3Client(
            minioSettings.AccessKey,
            minioSettings.SecretKey,
            new AmazonS3Config
            {
                ServiceURL = $"{(minioSettings.UseSSL ? "https" : "http")}://{minioSettings.Endpoint}",
                ForcePathStyle = true,
            }));
        services.AddScoped<IImageStorageService, MinioImageStorageService>();

        // RabbitMQ Event Publisher
        services.Configure<RabbitMqSettings>(configuration.GetSection("RabbitMq"));
        services.AddSingleton<ICatalogEventPublisher, RabbitMqCatalogEventPublisher>();

        // Handlers — Products
        services.AddScoped<CreateProductHandler>();
        services.AddScoped<UpdateProductHandler>();
        services.AddScoped<DeleteProductHandler>();
        services.AddScoped<ChangeProductStatusHandler>();
        services.AddScoped<SearchProductsHandler>();
        services.AddScoped<GetProductByIdHandler>();
        services.AddScoped<GetProductBySlugHandler>();
        services.AddScoped<UploadProductImageHandler>();
        services.AddScoped<DeleteProductImageHandler>();
        services.AddScoped<ReorderProductImagesHandler>();

        // Handlers — Categories
        services.AddScoped<CreateCategoryHandler>();
        services.AddScoped<UpdateCategoryHandler>();
        services.AddScoped<DeleteCategoryHandler>();
        services.AddScoped<GetCategoryTreeHandler>();

        // Handlers — Brands
        services.AddScoped<CreateBrandHandler>();
        services.AddScoped<UpdateBrandHandler>();
        services.AddScoped<DeleteBrandHandler>();

        // Health Checks
        services.AddHealthChecks()
            .AddMongoDb(sp => sp.GetRequiredService<IMongoDatabase>())
            .AddRedis(redisConnectionString)
            .AddRabbitMQ(sp =>
            {
                var rabbitUri = new Uri(configuration["RabbitMq:ConnectionString"] ?? "amqp://guest:guest@localhost:5672");
                var factory = new ConnectionFactory { Uri = rabbitUri };
                return factory.CreateConnectionAsync();
            });
    }
}

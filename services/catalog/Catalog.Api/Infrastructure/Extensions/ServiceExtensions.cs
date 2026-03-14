using Catalog.Api.Application.Handlers;
using Catalog.Api.Domain.Interfaces;
using Catalog.Api.Infrastructure.Repositories;
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

        // JWT Authentication
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
                        await ctx.Response.WriteAsync(
                            """{"code":"UNAUTHORIZED","message":"Authentication is required."}""");
                    },
                    OnForbidden = async ctx =>
                    {
                        ctx.Response.StatusCode = 403;
                        ctx.Response.ContentType = "application/json";
                        await ctx.Response.WriteAsync(
                            """{"code":"FORBIDDEN","message":"You do not have permission to access this resource."}""");
                    },
                };
            });
    }
}

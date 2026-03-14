using System.Net;
using System.Text.Json;
using Catalog.Api.Shared.Exceptions;
using MongoDB.Driver;

namespace Catalog.Api.Middlewares;

/// <summary>
/// Catches domain exceptions and translates them into structured JSON error responses.
/// </summary>
public sealed class ExceptionHandlingMiddleware(RequestDelegate next, ILogger<ExceptionHandlingMiddleware> logger)
{
    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNamingPolicy = JsonNamingPolicy.CamelCase,
    };

    /// <summary>
    /// Invokes the next middleware and handles any unhandled exceptions.
    /// </summary>
    public async Task InvokeAsync(HttpContext context)
    {
        try
        {
            await next(context);
        }
        catch (NotFoundException ex)
        {
            logger.LogWarning(ex, "Not found: {Code} - {Message}", ex.Code, ex.Message);
            await WriteErrorResponseAsync(context, HttpStatusCode.NotFound, ex.Code, ex.Message);
        }
        catch (ConflictException ex)
        {
            logger.LogWarning(ex, "Conflict: {Code} - {Message}", ex.Code, ex.Message);
            await WriteErrorResponseAsync(context, HttpStatusCode.Conflict, ex.Code, ex.Message);
        }
        catch (DomainException ex)
        {
            logger.LogWarning(ex, "Domain error: {Code} - {Message}", ex.Code, ex.Message);
            await WriteErrorResponseAsync(context, HttpStatusCode.BadRequest, ex.Code, ex.Message);
        }
        catch (MongoException ex)
        {
            logger.LogError(ex, "MongoDB error");
            await WriteErrorResponseAsync(context, HttpStatusCode.ServiceUnavailable, "SERVICE_UNAVAILABLE", "Service temporarily unavailable.");
        }
        catch (Exception ex)
        {
            logger.LogError(ex, "Unhandled exception");
            await WriteErrorResponseAsync(context, HttpStatusCode.InternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.");
        }
    }

    private static async Task WriteErrorResponseAsync(HttpContext context, HttpStatusCode statusCode, string code, string message)
    {
        context.Response.StatusCode = (int)statusCode;
        context.Response.ContentType = "application/json";

        var body = JsonSerializer.Serialize(new { code, message }, JsonOptions);
        await context.Response.WriteAsync(body);
    }
}

using System.Net.Http.Headers;
using System.Text;

namespace Catalog.Api.Middlewares;

public class SwaggerBasicAuthMiddleware(RequestDelegate next, IConfiguration configuration)
{
    public async Task InvokeAsync(HttpContext context)
    {
        var path = context.Request.Path;
        if (!path.StartsWithSegments("/swagger")
            || path.Value!.EndsWith(".js")
            || path.Value.EndsWith(".css")
            || path.Value.EndsWith(".png")
            || path.Value.EndsWith(".json"))
        {
            await next(context);
            return;
        }

        if (!context.Request.Headers.ContainsKey("Authorization"))
        {
            context.Response.StatusCode = 401;
            context.Response.Headers.WWWAuthenticate = "Basic realm=\"Swagger\"";
            return;
        }

        var authHeader = AuthenticationHeaderValue.Parse(context.Request.Headers.Authorization!);

        if (authHeader.Scheme != "Basic" || authHeader.Parameter is null)
        {
            context.Response.StatusCode = 401;
            context.Response.Headers.WWWAuthenticate = "Basic realm=\"Swagger\"";
            return;
        }

        var credentials = Encoding.UTF8.GetString(Convert.FromBase64String(authHeader.Parameter)).Split(':', 2);
        var username = credentials[0];
        var password = credentials.Length > 1 ? credentials[1] : "";

        var expectedUsername = configuration["Swagger:Username"] ?? "admin";
        var expectedPassword = configuration["Swagger:Password"] ?? "admin";

        if (username != expectedUsername || password != expectedPassword)
        {
            context.Response.StatusCode = 401;
            context.Response.Headers.WWWAuthenticate = "Basic realm=\"Swagger\"";
            return;
        }

        await next(context);
    }
}

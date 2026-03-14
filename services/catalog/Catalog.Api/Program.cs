using Catalog.Api.Infrastructure.Extensions;
using Catalog.Api.Middlewares;

var builder = WebApplication.CreateBuilder(args);

// * Swagger
builder.Services.AddSwagger();

builder.Services.AddAuthorization();

builder.Services.AddControllers();

// * Infrastructure
builder.Services.AddInfrastructure(builder.Configuration);

var app = builder.Build();

app.UseMiddleware<ExceptionHandlingMiddleware>();

// Configure the HTTP request pipeline.
if (app.Environment.IsDevelopment())
{
    app.UseSwagger();
    app.UseSwaggerUI();
}

app.UseHttpsRedirection();
app.UseAuthentication();
app.UseAuthorization();
app.MapControllers();

await app.RunAsync();

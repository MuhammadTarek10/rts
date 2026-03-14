using Catalog.Api.Application.DTOs;
using Catalog.Api.Domain.Entities;
using Catalog.Api.Domain.Events;
using Catalog.Api.Infrastructure;
using Catalog.Api.Infrastructure.Messaging;
using Catalog.Api.Infrastructure.Services;
using Catalog.Api.Shared.Exceptions;

namespace Catalog.Api.Application.Handlers;

public sealed class CreateBrandHandler(
    CatalogUnitOfWork uow,
    ICacheService cache,
    ICatalogEventPublisher eventPublisher,
    ILogger<CreateBrandHandler> logger)
{
    public async Task<BrandResponseDto> HandleAsync(
        CreateBrandDto request,
        CancellationToken cancellationToken = default)
    {
        logger.LogInformation("Creating brand {Name}", request.Name);

        if (await uow.Brands.ExistsBySlugAsync(request.Slug, cancellationToken))
            throw new ConflictException("DUPLICATE_SLUG", "A brand with this slug already exists.");

        var brand = new Brand
        {
            Name = request.Name.Trim(),
            Slug = request.Slug.Trim().ToLowerInvariant(),
            Description = request.Description?.Trim(),
            LogoUrl = request.LogoUrl,
            Website = request.Website,
        };

        await uow.Brands.SaveAsync(brand, isNew: true, cancellationToken);

        await cache.RemoveAsync("catalog:brands:all", cancellationToken);

        await eventPublisher.PublishAsync(new BrandCreatedEvent(brand.Id, brand.Name), cancellationToken);

        logger.LogInformation("Brand {BrandId} created: {Name}", brand.Id, brand.Name);
        return BrandResponseDto.FromEntity(brand);
    }
}

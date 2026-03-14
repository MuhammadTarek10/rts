using Catalog.Api.Application.DTOs;
using Catalog.Api.Application.Handlers;
using Catalog.Api.Domain.Entities;
using Catalog.Api.Domain.Interfaces;
using Catalog.Api.Infrastructure;
using Catalog.Api.Infrastructure.Messaging;
using Catalog.Api.Infrastructure.Services;
using Catalog.Api.Shared.Abstraction;
using Catalog.Api.Shared.Exceptions;
using FluentAssertions;
using Microsoft.Extensions.Logging.Abstractions;
using Moq;

namespace Catalog.Api.Tests;

public class BrandHandlerTests
{
    private readonly Mock<IProductRepository> _productRepo = new();
    private readonly Mock<ICategoryRepository> _categoryRepo = new();
    private readonly Mock<IBrandRepository> _brandRepo = new();
    private readonly Mock<ICacheService> _cache = new();
    private readonly Mock<ICatalogEventPublisher> _events = new();
    private readonly CatalogUnitOfWork _uow;

    public BrandHandlerTests()
    {
        _uow = new CatalogUnitOfWork(_productRepo.Object, _categoryRepo.Object, _brandRepo.Object);
        _cache.Setup(c => c.RemoveAsync(It.IsAny<string>(), It.IsAny<CancellationToken>())).Returns(Task.CompletedTask);
        _events.Setup(e => e.PublishAsync(It.IsAny<IDomainEvent>(), It.IsAny<CancellationToken>())).Returns(Task.CompletedTask);
    }

    [Fact]
    public async Task CreateBrand_Success_ReturnsDto()
    {
        _brandRepo.Setup(r => r.ExistsBySlugAsync("brand-slug", default)).ReturnsAsync(false);
        _brandRepo.Setup(r => r.SaveAsync(It.IsAny<Brand>(), true, default)).Returns(Task.CompletedTask);

        var handler = new CreateBrandHandler(_uow, _cache.Object, _events.Object, NullLogger<CreateBrandHandler>.Instance);
        var result = await handler.HandleAsync(new CreateBrandDto { Name = "Nike", Slug = "brand-slug" });

        result.Should().NotBeNull();
        result.Name.Should().Be("Nike");
        result.IsActive.Should().BeTrue();
    }

    [Fact]
    public async Task CreateBrand_DuplicateSlug_ThrowsConflict()
    {
        _brandRepo.Setup(r => r.ExistsBySlugAsync("dup", default)).ReturnsAsync(true);
        var handler = new CreateBrandHandler(_uow, _cache.Object, _events.Object, NullLogger<CreateBrandHandler>.Instance);
        var act = async () => await handler.HandleAsync(new CreateBrandDto { Name = "N", Slug = "dup" });
        await act.Should().ThrowAsync<ConflictException>();
    }

    [Fact]
    public async Task CreateBrand_SlugNormalized_ToLowercase()
    {
        _brandRepo.Setup(r => r.ExistsBySlugAsync("nike-brand", default)).ReturnsAsync(false);
        _brandRepo.Setup(r => r.SaveAsync(It.IsAny<Brand>(), true, default)).Returns(Task.CompletedTask);

        var handler = new CreateBrandHandler(_uow, _cache.Object, _events.Object, NullLogger<CreateBrandHandler>.Instance);
        var result = await handler.HandleAsync(new CreateBrandDto { Name = "Nike", Slug = "Nike-Brand" });

        result.Slug.Should().Be("nike-brand");
    }

    [Fact]
    public async Task DeleteBrand_HasProducts_ThrowsConflict()
    {
        var brand = new Brand { Name = "B", Slug = "b" };
        _brandRepo.Setup(r => r.GetByIdAsync(brand.Id, default)).ReturnsAsync(brand);
        _brandRepo.Setup(r => r.HasProductsAsync(brand.Id, default)).ReturnsAsync(true);

        var handler = new DeleteBrandHandler(_uow, _cache.Object, NullLogger<DeleteBrandHandler>.Instance);
        var act = async () => await handler.HandleAsync(brand.Id);
        var exception = await act.Should().ThrowAsync<ConflictException>();
        exception.WithMessage("*products*");
    }

    [Fact]
    public async Task DeleteBrand_Success_Deletes()
    {
        var brand = new Brand { Name = "B2", Slug = "b2" };
        _brandRepo.Setup(r => r.GetByIdAsync(brand.Id, default)).ReturnsAsync(brand);
        _brandRepo.Setup(r => r.HasProductsAsync(brand.Id, default)).ReturnsAsync(false);
        _brandRepo.Setup(r => r.DeleteAsync(brand.Id, default)).Returns(Task.CompletedTask);

        var handler = new DeleteBrandHandler(_uow, _cache.Object, NullLogger<DeleteBrandHandler>.Instance);
        var act = async () => await handler.HandleAsync(brand.Id);
        await act.Should().NotThrowAsync();
        _brandRepo.Verify(r => r.DeleteAsync(brand.Id, default), Times.Once);
    }

    [Fact]
    public async Task DeleteBrand_NotFound_ThrowsNotFoundException()
    {
        _brandRepo.Setup(r => r.GetByIdAsync("missing", default)).ReturnsAsync((Brand?)null);
        var handler = new DeleteBrandHandler(_uow, _cache.Object, NullLogger<DeleteBrandHandler>.Instance);
        var act = async () => await handler.HandleAsync("missing");
        await act.Should().ThrowAsync<NotFoundException>();
    }
}

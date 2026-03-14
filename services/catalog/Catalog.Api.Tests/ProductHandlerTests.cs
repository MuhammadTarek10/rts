using Catalog.Api.Application.DTOs;
using Catalog.Api.Application.Handlers;
using Catalog.Api.Domain.Entities;
using Catalog.Api.Domain.Models;
using Catalog.Api.Infrastructure;
using Catalog.Api.Infrastructure.Messaging;
using Catalog.Api.Infrastructure.Services;
using Catalog.Api.Domain.Interfaces;
using Catalog.Api.Shared.Abstraction;
using Catalog.Api.Shared.Exceptions;
using FluentAssertions;
using Microsoft.Extensions.Logging.Abstractions;
using Moq;

namespace Catalog.Api.Tests;

public class ProductHandlerTests
{
    private readonly Mock<IProductRepository> _productRepo = new();
    private readonly Mock<ICategoryRepository> _categoryRepo = new();
    private readonly Mock<IBrandRepository> _brandRepo = new();
    private readonly Mock<ICacheService> _cache = new();
    private readonly Mock<ICatalogEventPublisher> _events = new();
    private readonly CatalogUnitOfWork _uow;

    public ProductHandlerTests()
    {
        _uow = new CatalogUnitOfWork(_productRepo.Object, _categoryRepo.Object, _brandRepo.Object);
        // Default: cache miss
        _cache.Setup(c => c.GetAsync<SearchProductsResponse>(It.IsAny<string>(), It.IsAny<CancellationToken>()))
              .ReturnsAsync((SearchProductsResponse?)null);
        _cache.Setup(c => c.SetAsync(It.IsAny<string>(), It.IsAny<SearchProductsResponse>(), It.IsAny<TimeSpan>(), It.IsAny<CancellationToken>()))
              .Returns(Task.CompletedTask);
        _cache.Setup(c => c.RemoveAsync(It.IsAny<string>(), It.IsAny<CancellationToken>()))
              .Returns(Task.CompletedTask);
        _events.Setup(e => e.PublishAsync(It.IsAny<IDomainEvent>(), It.IsAny<CancellationToken>()))
               .Returns(Task.CompletedTask);
    }

    private CreateProductHandler CreateHandler() =>
        new(_uow, _events.Object, NullLogger<CreateProductHandler>.Instance);

    // --- CreateProduct Tests ---

    [Fact]
    public async Task CreateProduct_Success_ReturnsDto()
    {
        _productRepo.Setup(r => r.ExistsBySkuAsync("SKU1", default)).ReturnsAsync(false);
        _productRepo.Setup(r => r.ExistsBySlugAsync("sku-1", default)).ReturnsAsync(false);
        _brandRepo.Setup(r => r.ExistsAsync("brand1", default)).ReturnsAsync(true);
        _categoryRepo.Setup(r => r.GetByIdAsync("cat1", default)).ReturnsAsync(new Category { Name = "Cat", Slug = "cat" });
        _productRepo.Setup(r => r.SaveAsync(It.IsAny<Product>(), true, default)).Returns(Task.CompletedTask);

        var dto = new CreateProductDto { Sku = "SKU1", Slug = "sku-1", Title = "Test", Amount = 10m, Currency = "USD", BrandId = "brand1", CategoryIds = ["cat1"] };
        var result = await CreateHandler().HandleAsync(dto);

        result.Should().NotBeNull();
        result.Sku.Should().Be("SKU1");
        result.Status.Should().Be("Draft");
    }

    [Fact]
    public async Task CreateProduct_DuplicateSku_ThrowsConflict()
    {
        _productRepo.Setup(r => r.ExistsBySkuAsync("SKU1", default)).ReturnsAsync(true);
        var dto = new CreateProductDto { Sku = "SKU1", Slug = "sku-1", Title = "Test", Amount = 10m, Currency = "USD" };
        var act = async () => await CreateHandler().HandleAsync(dto);
        var exception = await act.Should().ThrowAsync<ConflictException>();
        exception.WithMessage("*SKU*");
    }

    [Fact]
    public async Task CreateProduct_DuplicateSlug_ThrowsConflict()
    {
        _productRepo.Setup(r => r.ExistsBySkuAsync("SKU2", default)).ReturnsAsync(false);
        _productRepo.Setup(r => r.ExistsBySlugAsync("slug-2", default)).ReturnsAsync(true);
        var dto = new CreateProductDto { Sku = "SKU2", Slug = "slug-2", Title = "Test", Amount = 10m, Currency = "USD" };
        var act = async () => await CreateHandler().HandleAsync(dto);
        await act.Should().ThrowAsync<ConflictException>();
    }

    [Fact]
    public async Task CreateProduct_InvalidBrandId_ThrowsDomainException()
    {
        _productRepo.Setup(r => r.ExistsBySkuAsync("SKU3", default)).ReturnsAsync(false);
        _productRepo.Setup(r => r.ExistsBySlugAsync("sku-3", default)).ReturnsAsync(false);
        _brandRepo.Setup(r => r.ExistsAsync("bad-brand", default)).ReturnsAsync(false);
        var dto = new CreateProductDto { Sku = "SKU3", Slug = "sku-3", Title = "Test", Amount = 10m, Currency = "USD", BrandId = "bad-brand" };
        var act = async () => await CreateHandler().HandleAsync(dto);
        await act.Should().ThrowAsync<DomainException>();
    }

    [Fact]
    public async Task CreateProduct_InvalidCategoryId_ThrowsDomainException()
    {
        _productRepo.Setup(r => r.ExistsBySkuAsync("SKU4", default)).ReturnsAsync(false);
        _productRepo.Setup(r => r.ExistsBySlugAsync("sku-4", default)).ReturnsAsync(false);
        _categoryRepo.Setup(r => r.GetByIdAsync("bad-cat", default)).ReturnsAsync((Category?)null);
        var dto = new CreateProductDto { Sku = "SKU4", Slug = "sku-4", Title = "Test", Amount = 10m, Currency = "USD", CategoryIds = ["bad-cat"] };
        var act = async () => await CreateHandler().HandleAsync(dto);
        await act.Should().ThrowAsync<DomainException>();
    }

    [Fact]
    public async Task CreateProduct_NoCategories_NoBrandId_Succeeds()
    {
        _productRepo.Setup(r => r.ExistsBySkuAsync("SKU5", default)).ReturnsAsync(false);
        _productRepo.Setup(r => r.ExistsBySlugAsync("sku-5", default)).ReturnsAsync(false);
        _productRepo.Setup(r => r.SaveAsync(It.IsAny<Product>(), true, default)).Returns(Task.CompletedTask);

        var dto = new CreateProductDto { Sku = "SKU5", Slug = "sku-5", Title = "Minimal", Amount = 1m, Currency = "USD" };
        var result = await CreateHandler().HandleAsync(dto);

        result.BrandId.Should().BeNull();
        result.CategoryIds.Should().BeEmpty();
    }

    // --- UpdateProduct Tests ---

    [Fact]
    public async Task UpdateProduct_Success_ReturnsUpdatedDto()
    {
        var product = new Product { Sku = "SKU5", Slug = "sku-5", Title = "Old", Price = new Money { Amount = 5m, Currency = "USD" } };
        _productRepo.Setup(r => r.GetByIdAsync("p1", default)).ReturnsAsync(product);
        _brandRepo.Setup(r => r.ExistsAsync(It.IsAny<string>(), default)).ReturnsAsync(true);
        _productRepo.Setup(r => r.ReplaceIfVersionMatchesAsync(It.IsAny<Product>(), It.IsAny<int>(), default)).Returns(Task.CompletedTask);

        var handler = new UpdateProductHandler(_uow, _cache.Object, _events.Object, NullLogger<UpdateProductHandler>.Instance);
        var dto = new UpdateProductDto { Title = "New", Amount = 20m, Currency = "USD", Version = 1 };
        var result = await handler.HandleAsync("p1", dto);

        result.Title.Should().Be("New");
        result.Amount.Should().Be(20m);
    }

    [Fact]
    public async Task UpdateProduct_NotFound_ThrowsNotFoundException()
    {
        _productRepo.Setup(r => r.GetByIdAsync("missing", default)).ReturnsAsync((Product?)null);
        var handler = new UpdateProductHandler(_uow, _cache.Object, _events.Object, NullLogger<UpdateProductHandler>.Instance);
        var act = async () => await handler.HandleAsync("missing", new UpdateProductDto { Title = "X", Amount = 1, Currency = "USD" });
        await act.Should().ThrowAsync<NotFoundException>();
    }

    [Fact]
    public async Task UpdateProduct_StaleVersion_ThrowsConflict()
    {
        var product = new Product { Sku = "SKU6", Slug = "sku-6", Title = "T", Price = new Money { Amount = 1m, Currency = "USD" } };
        _productRepo.Setup(r => r.GetByIdAsync("p2", default)).ReturnsAsync(product);
        _productRepo.Setup(r => r.ReplaceIfVersionMatchesAsync(It.IsAny<Product>(), It.IsAny<int>(), default))
            .ThrowsAsync(new ConflictException("VERSION_CONFLICT", "Modified by another user"));
        var handler = new UpdateProductHandler(_uow, _cache.Object, _events.Object, NullLogger<UpdateProductHandler>.Instance);
        var act = async () => await handler.HandleAsync("p2", new UpdateProductDto { Title = "X", Amount = 1, Currency = "USD" });
        await act.Should().ThrowAsync<ConflictException>();
    }

    [Fact]
    public async Task UpdateProduct_InvalidBrandId_ThrowsDomainException()
    {
        var product = new Product { Sku = "SKU7", Slug = "sku-7", Title = "T", Price = new Money { Amount = 1m, Currency = "USD" } };
        _productRepo.Setup(r => r.GetByIdAsync("p3", default)).ReturnsAsync(product);
        _brandRepo.Setup(r => r.ExistsAsync("bad-brand", default)).ReturnsAsync(false);
        var handler = new UpdateProductHandler(_uow, _cache.Object, _events.Object, NullLogger<UpdateProductHandler>.Instance);
        var act = async () => await handler.HandleAsync("p3", new UpdateProductDto { Title = "X", Amount = 1, Currency = "USD", BrandId = "bad-brand" });
        await act.Should().ThrowAsync<DomainException>();
    }

    // --- DeleteProduct Tests ---

    [Fact]
    public async Task DeleteProduct_Success_ArchivesProduct()
    {
        var product = new Product { Sku = "SKU8", Slug = "sku-8", Title = "T", Price = new Money { Amount = 1m, Currency = "USD" } };
        _productRepo.Setup(r => r.GetByIdAsync("p4", default)).ReturnsAsync(product);
        _productRepo.Setup(r => r.UpdateAsync(It.IsAny<Product>(), default)).Returns(Task.CompletedTask);
        var handler = new DeleteProductHandler(_uow, _cache.Object, _events.Object, NullLogger<DeleteProductHandler>.Instance);
        await handler.HandleAsync("p4");
        product.Status.Should().Be(ProductStatus.Archived);
    }

    [Fact]
    public async Task DeleteProduct_NotFound_ThrowsNotFoundException()
    {
        _productRepo.Setup(r => r.GetByIdAsync("x", default)).ReturnsAsync((Product?)null);
        var handler = new DeleteProductHandler(_uow, _cache.Object, _events.Object, NullLogger<DeleteProductHandler>.Instance);
        var act = async () => await handler.HandleAsync("x");
        await act.Should().ThrowAsync<NotFoundException>();
    }

    // --- ChangeStatus Tests ---

    [Fact]
    public async Task ChangeStatus_DraftToActive_Success()
    {
        var product = new Product { Sku = "SKU9", Slug = "sku-9", Title = "T", Price = new Money { Amount = 1m, Currency = "USD" } };
        _productRepo.Setup(r => r.GetByIdAsync("p5", default)).ReturnsAsync(product);
        _productRepo.Setup(r => r.UpdateAsync(It.IsAny<Product>(), default)).Returns(Task.CompletedTask);
        var handler = new ChangeProductStatusHandler(_uow, _cache.Object, _events.Object, NullLogger<ChangeProductStatusHandler>.Instance);
        var result = await handler.HandleAsync("p5", new ChangeStatusDto { Status = "Active" });
        result.Status.Should().Be("Active");
    }

    [Fact]
    public async Task ChangeStatus_ActiveToArchived_Success()
    {
        var product = new Product { Sku = "SKU10", Slug = "sku-10", Title = "T", Price = new Money { Amount = 1m, Currency = "USD" } };
        product.SetStatus(ProductStatus.Active);
        _productRepo.Setup(r => r.GetByIdAsync("p6", default)).ReturnsAsync(product);
        _productRepo.Setup(r => r.UpdateAsync(It.IsAny<Product>(), default)).Returns(Task.CompletedTask);
        var handler = new ChangeProductStatusHandler(_uow, _cache.Object, _events.Object, NullLogger<ChangeProductStatusHandler>.Instance);
        var result = await handler.HandleAsync("p6", new ChangeStatusDto { Status = "Archived" });
        result.Status.Should().Be("Archived");
    }

    [Fact]
    public async Task ChangeStatus_NotFound_ThrowsNotFoundException()
    {
        _productRepo.Setup(r => r.GetByIdAsync("y", default)).ReturnsAsync((Product?)null);
        var handler = new ChangeProductStatusHandler(_uow, _cache.Object, _events.Object, NullLogger<ChangeProductStatusHandler>.Instance);
        var act = async () => await handler.HandleAsync("y", new ChangeStatusDto { Status = "Active" });
        await act.Should().ThrowAsync<NotFoundException>();
    }

    [Fact]
    public async Task ChangeStatus_InvalidStatus_ThrowsDomainException()
    {
        var product = new Product { Sku = "SKU11", Slug = "sku-11", Title = "T", Price = new Money { Amount = 1m, Currency = "USD" } };
        _productRepo.Setup(r => r.GetByIdAsync("p7", default)).ReturnsAsync(product);
        var handler = new ChangeProductStatusHandler(_uow, _cache.Object, _events.Object, NullLogger<ChangeProductStatusHandler>.Instance);
        var act = async () => await handler.HandleAsync("p7", new ChangeStatusDto { Status = "InvalidStatus" });
        await act.Should().ThrowAsync<DomainException>();
    }

    // --- SearchProducts Tests ---

    [Fact]
    public async Task SearchProducts_TextQuery_ReturnsMatches()
    {
        var products = new List<Product> { new Product { Sku = "S1", Slug = "s1", Title = "Shirt", Price = new Money { Amount = 25m, Currency = "USD" } } };
        _productRepo.Setup(r => r.SearchAsync(It.IsAny<ProductSearchCriteria>(), default))
            .ReturnsAsync(((IReadOnlyList<Product>)products, 1L));
        _categoryRepo.Setup(r => r.GetDescendantIdsAsync(It.IsAny<string>(), default))
            .ReturnsAsync(new List<string>());

        var handler = new SearchProductsHandler(_uow, _cache.Object, NullLogger<SearchProductsHandler>.Instance);
        var result = await handler.HandleAsync(new SearchProductsRequest { Query = "shirt" });

        result.Items.Should().HaveCount(1);
        result.TotalCount.Should().Be(1);
    }

    [Fact]
    public async Task SearchProducts_CategoryFilter_IncludesDescendants()
    {
        _categoryRepo.Setup(r => r.GetDescendantIdsAsync("cat1", default))
            .ReturnsAsync(new List<string> { "cat2", "cat3" });

        ProductSearchCriteria? capturedCriteria = null;
        _productRepo.Setup(r => r.SearchAsync(It.IsAny<ProductSearchCriteria>(), default))
            .Callback<ProductSearchCriteria, CancellationToken>((c, _) => capturedCriteria = c)
            .ReturnsAsync(((IReadOnlyList<Product>)new List<Product>(), 0L));

        var handler = new SearchProductsHandler(_uow, _cache.Object, NullLogger<SearchProductsHandler>.Instance);
        await handler.HandleAsync(new SearchProductsRequest { CategoryId = "cat1" });

        capturedCriteria!.CategoryIds.Should().Contain("cat1");
        capturedCriteria.CategoryIds.Should().Contain("cat2");
        capturedCriteria.CategoryIds.Should().Contain("cat3");
    }

    [Fact]
    public async Task SearchProducts_EmptyResults_ReturnsEmptyList()
    {
        _productRepo.Setup(r => r.SearchAsync(It.IsAny<ProductSearchCriteria>(), default))
            .ReturnsAsync(((IReadOnlyList<Product>)new List<Product>(), 0L));

        var handler = new SearchProductsHandler(_uow, _cache.Object, NullLogger<SearchProductsHandler>.Instance);
        var result = await handler.HandleAsync(new SearchProductsRequest { });

        result.Items.Should().BeEmpty();
        result.TotalCount.Should().Be(0);
    }

    [Fact]
    public async Task SearchProducts_PageSizeClampedToMax100()
    {
        ProductSearchCriteria? captured = null;
        _productRepo.Setup(r => r.SearchAsync(It.IsAny<ProductSearchCriteria>(), default))
            .Callback<ProductSearchCriteria, CancellationToken>((c, _) => captured = c)
            .ReturnsAsync(((IReadOnlyList<Product>)new List<Product>(), 0L));

        var handler = new SearchProductsHandler(_uow, _cache.Object, NullLogger<SearchProductsHandler>.Instance);
        await handler.HandleAsync(new SearchProductsRequest { PageSize = 999 });

        captured!.PageSize.Should().Be(100);
    }
}

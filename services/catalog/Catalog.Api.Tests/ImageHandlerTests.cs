using Catalog.Api.Application.DTOs;
using Catalog.Api.Application.Handlers;
using Catalog.Api.Domain.Entities;
using Catalog.Api.Domain.Interfaces;
using Catalog.Api.Infrastructure;
using Catalog.Api.Infrastructure.Services;
using Catalog.Api.Shared.Exceptions;
using FluentAssertions;
using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.Logging.Abstractions;
using Moq;

namespace Catalog.Api.Tests;

public class ImageHandlerTests
{
    private readonly Mock<IProductRepository> _productRepo = new();
    private readonly Mock<ICategoryRepository> _categoryRepo = new();
    private readonly Mock<IBrandRepository> _brandRepo = new();
    private readonly Mock<ICacheService> _cache = new();
    private readonly Mock<IImageStorageService> _imageStorage = new();
    private readonly CatalogUnitOfWork _uow;

    public ImageHandlerTests()
    {
        _uow = new CatalogUnitOfWork(_productRepo.Object, _categoryRepo.Object, _brandRepo.Object);
        _cache.Setup(c => c.RemoveAsync(It.IsAny<string>(), It.IsAny<CancellationToken>())).Returns(Task.CompletedTask);
    }

    private IFormFile CreateMockFile(string contentType, byte[] content, string fileName = "test.jpg")
    {
        var stream = new MemoryStream(content);
        var file = new Mock<IFormFile>();
        file.Setup(f => f.ContentType).Returns(contentType);
        file.Setup(f => f.Length).Returns(content.Length);
        file.Setup(f => f.FileName).Returns(fileName);
        file.Setup(f => f.OpenReadStream()).Returns(stream);
        return file.Object;
    }

    [Fact]
    public async Task UploadImage_Success_ImageAddedWithCorrectSortOrder()
    {
        var existingImage = new ProductImage { Url = "http://old.jpg", Key = "key1", SortOrder = 0, IsPrimary = true };
        var product = new Product
        {
            Sku = "S1", Slug = "s1", Title = "T",
            Price = new Money { Amount = 1m, Currency = "USD" },
            Images = [existingImage]
        };
        _productRepo.Setup(r => r.GetByIdAsync("p1", default)).ReturnsAsync(product);
        _productRepo.Setup(r => r.UpdateAsync(It.IsAny<Product>(), default)).Returns(Task.CompletedTask);
        _imageStorage.Setup(s => s.UploadAsync(It.IsAny<Stream>(), It.IsAny<string>(), It.IsAny<string>(), "p1", default))
            .ReturnsAsync(("http://new.jpg", "key2"));

        // Valid JPEG magic bytes
        var jpegContent = new byte[] { 0xFF, 0xD8, 0xFF, 0x00, 0x01 };
        var file = CreateMockFile("image/jpeg", jpegContent);

        var handler = new UploadProductImageHandler(_uow, _imageStorage.Object, _cache.Object, NullLogger<UploadProductImageHandler>.Instance);
        var result = await handler.HandleAsync("p1", file, null);

        result.Images.Should().HaveCount(2);
        result.Images.Last().SortOrder.Should().Be(1); // after existing image with SortOrder=0
        result.Images.Last().IsPrimary.Should().BeFalse(); // not primary since another exists
    }

    [Fact]
    public async Task UploadImage_FirstImage_IsPrimary()
    {
        var product = new Product
        {
            Sku = "S0", Slug = "s0", Title = "T",
            Price = new Money { Amount = 1m, Currency = "USD" },
            Images = []
        };
        _productRepo.Setup(r => r.GetByIdAsync("p0", default)).ReturnsAsync(product);
        _productRepo.Setup(r => r.UpdateAsync(It.IsAny<Product>(), default)).Returns(Task.CompletedTask);
        _imageStorage.Setup(s => s.UploadAsync(It.IsAny<Stream>(), It.IsAny<string>(), It.IsAny<string>(), "p0", default))
            .ReturnsAsync(("http://first.jpg", "first-key"));

        var jpegContent = new byte[] { 0xFF, 0xD8, 0xFF, 0x00 };
        var file = CreateMockFile("image/jpeg", jpegContent);

        var handler = new UploadProductImageHandler(_uow, _imageStorage.Object, _cache.Object, NullLogger<UploadProductImageHandler>.Instance);
        var result = await handler.HandleAsync("p0", file, null);

        result.Images.Should().HaveCount(1);
        result.Images[0].IsPrimary.Should().BeTrue();
        result.Images[0].SortOrder.Should().Be(0);
    }

    [Fact]
    public async Task UploadImage_FileTooLarge_ThrowsDomainException()
    {
        // File size check happens before product lookup
        var bigContent = new byte[6 * 1024 * 1024]; // 6MB
        var file = CreateMockFile("image/jpeg", bigContent);

        var handler = new UploadProductImageHandler(_uow, _imageStorage.Object, _cache.Object, NullLogger<UploadProductImageHandler>.Instance);
        var act = async () => await handler.HandleAsync("p2", file, null);
        var exception = await act.Should().ThrowAsync<DomainException>();
        exception.WithMessage("*5MB*");
    }

    [Fact]
    public async Task UploadImage_InvalidContentType_ThrowsDomainException()
    {
        // Content type check happens before product lookup
        var file = CreateMockFile("image/gif", new byte[] { 0x47, 0x49, 0x46, 0x38 }); // GIF bytes

        var handler = new UploadProductImageHandler(_uow, _imageStorage.Object, _cache.Object, NullLogger<UploadProductImageHandler>.Instance);
        var act = async () => await handler.HandleAsync("p3", file, null);
        await act.Should().ThrowAsync<DomainException>();
    }

    [Fact]
    public async Task UploadImage_ArchivedProduct_ThrowsDomainException()
    {
        var product = new Product { Sku = "S4", Slug = "s4", Title = "T", Price = new Money { Amount = 1m, Currency = "USD" } };
        product.SetStatus(ProductStatus.Archived);
        _productRepo.Setup(r => r.GetByIdAsync("p4", default)).ReturnsAsync(product);

        // Must pass content-type check (jpeg) and size check to reach the archived check
        var jpegContent = new byte[] { 0xFF, 0xD8, 0xFF, 0x00 };
        var file = CreateMockFile("image/jpeg", jpegContent);

        var handler = new UploadProductImageHandler(_uow, _imageStorage.Object, _cache.Object, NullLogger<UploadProductImageHandler>.Instance);
        var act = async () => await handler.HandleAsync("p4", file, null);
        var exception = await act.Should().ThrowAsync<DomainException>();
        exception.WithMessage("*archived*");
    }

    [Fact]
    public async Task UploadImage_ProductNotFound_ThrowsNotFoundException()
    {
        _productRepo.Setup(r => r.GetByIdAsync("missing", default)).ReturnsAsync((Product?)null);

        var jpegContent = new byte[] { 0xFF, 0xD8, 0xFF, 0x00 };
        var file = CreateMockFile("image/jpeg", jpegContent);

        var handler = new UploadProductImageHandler(_uow, _imageStorage.Object, _cache.Object, NullLogger<UploadProductImageHandler>.Instance);
        var act = async () => await handler.HandleAsync("missing", file, null);
        await act.Should().ThrowAsync<NotFoundException>();
    }

    [Fact]
    public async Task DeleteImage_Success_ImageRemoved()
    {
        var img = new ProductImage { ImageId = "img1", Url = "http://u.jpg", Key = "key1", SortOrder = 0, IsPrimary = true };
        var product = new Product { Sku = "S5", Slug = "s5", Title = "T", Price = new Money { Amount = 1m, Currency = "USD" }, Images = [img] };
        _productRepo.Setup(r => r.GetByIdAsync("p5", default)).ReturnsAsync(product);
        _productRepo.Setup(r => r.UpdateAsync(It.IsAny<Product>(), default)).Returns(Task.CompletedTask);
        _imageStorage.Setup(s => s.DeleteAsync("key1", default)).Returns(Task.CompletedTask);

        var handler = new DeleteProductImageHandler(_uow, _imageStorage.Object, _cache.Object, NullLogger<DeleteProductImageHandler>.Instance);
        var result = await handler.HandleAsync("p5", "img1");

        result.Images.Should().BeEmpty();
    }

    [Fact]
    public async Task DeleteImage_ImageNotFound_ThrowsNotFoundException()
    {
        var product = new Product { Sku = "S6", Slug = "s6", Title = "T", Price = new Money { Amount = 1m, Currency = "USD" }, Images = [] };
        _productRepo.Setup(r => r.GetByIdAsync("p6", default)).ReturnsAsync(product);

        var handler = new DeleteProductImageHandler(_uow, _imageStorage.Object, _cache.Object, NullLogger<DeleteProductImageHandler>.Instance);
        var act = async () => await handler.HandleAsync("p6", "non-existent-img");
        await act.Should().ThrowAsync<NotFoundException>();
    }

    [Fact]
    public async Task DeleteImage_PrimaryDeleted_PromotesNextImage()
    {
        var img1 = new ProductImage { ImageId = "img1", Url = "http://u1.jpg", Key = "k1", SortOrder = 0, IsPrimary = true };
        var img2 = new ProductImage { ImageId = "img2", Url = "http://u2.jpg", Key = "k2", SortOrder = 1, IsPrimary = false };
        var product = new Product { Sku = "S7p", Slug = "s7p", Title = "T", Price = new Money { Amount = 1m, Currency = "USD" }, Images = [img1, img2] };
        _productRepo.Setup(r => r.GetByIdAsync("p7p", default)).ReturnsAsync(product);
        _productRepo.Setup(r => r.UpdateAsync(It.IsAny<Product>(), default)).Returns(Task.CompletedTask);
        _imageStorage.Setup(s => s.DeleteAsync("k1", default)).Returns(Task.CompletedTask);

        var handler = new DeleteProductImageHandler(_uow, _imageStorage.Object, _cache.Object, NullLogger<DeleteProductImageHandler>.Instance);
        var result = await handler.HandleAsync("p7p", "img1");

        result.Images.Should().HaveCount(1);
        result.Images[0].IsPrimary.Should().BeTrue();
        result.Images[0].ImageId.Should().Be("img2");
    }

    [Fact]
    public async Task ReorderImages_Success_SortOrdersUpdated()
    {
        var img1 = new ProductImage { ImageId = "img1", Url = "http://u1.jpg", Key = "k1", SortOrder = 0 };
        var img2 = new ProductImage { ImageId = "img2", Url = "http://u2.jpg", Key = "k2", SortOrder = 1 };
        var product = new Product { Sku = "S7", Slug = "s7", Title = "T", Price = new Money { Amount = 1m, Currency = "USD" }, Images = [img1, img2] };
        _productRepo.Setup(r => r.GetByIdAsync("p7", default)).ReturnsAsync(product);
        _productRepo.Setup(r => r.UpdateAsync(It.IsAny<Product>(), default)).Returns(Task.CompletedTask);

        var handler = new ReorderProductImagesHandler(_uow, _cache.Object, NullLogger<ReorderProductImagesHandler>.Instance);
        var reorderDto = new ReorderImagesDto
        {
            Order = [new ImageOrderItem { ImageId = "img1", SortOrder = 5 }, new ImageOrderItem { ImageId = "img2", SortOrder = 3 }]
        };
        var result = await handler.HandleAsync("p7", reorderDto);

        result.Images.First().SortOrder.Should().Be(3); // img2 now first (sorted by SortOrder)
        result.Images.Last().SortOrder.Should().Be(5);  // img1 now last
    }

    [Fact]
    public async Task ReorderImages_ProductNotFound_ThrowsNotFoundException()
    {
        _productRepo.Setup(r => r.GetByIdAsync("missing", default)).ReturnsAsync((Product?)null);

        var handler = new ReorderProductImagesHandler(_uow, _cache.Object, NullLogger<ReorderProductImagesHandler>.Instance);
        var act = async () => await handler.HandleAsync("missing", new ReorderImagesDto());
        await act.Should().ThrowAsync<NotFoundException>();
    }
}

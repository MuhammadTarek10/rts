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

public class CategoryHandlerTests
{
    private readonly Mock<IProductRepository> _productRepo = new();
    private readonly Mock<ICategoryRepository> _categoryRepo = new();
    private readonly Mock<IBrandRepository> _brandRepo = new();
    private readonly Mock<ICacheService> _cache = new();
    private readonly Mock<ICatalogEventPublisher> _events = new();
    private readonly CatalogUnitOfWork _uow;

    public CategoryHandlerTests()
    {
        _uow = new CatalogUnitOfWork(_productRepo.Object, _categoryRepo.Object, _brandRepo.Object);
        _cache.Setup(c => c.RemoveAsync(It.IsAny<string>(), It.IsAny<CancellationToken>())).Returns(Task.CompletedTask);
        _cache.Setup(c => c.GetAsync<List<CategoryTreeDto>>(It.IsAny<string>(), It.IsAny<CancellationToken>())).ReturnsAsync((List<CategoryTreeDto>?)null);
        _cache.Setup(c => c.SetAsync(It.IsAny<string>(), It.IsAny<List<CategoryTreeDto>>(), It.IsAny<TimeSpan>(), It.IsAny<CancellationToken>())).Returns(Task.CompletedTask);
        _events.Setup(e => e.PublishAsync(It.IsAny<IDomainEvent>(), It.IsAny<CancellationToken>())).Returns(Task.CompletedTask);
    }

    [Fact]
    public async Task CreateCategory_Success_ReturnsDto()
    {
        _categoryRepo.Setup(r => r.ExistsBySlugAsync("cat-slug", default)).ReturnsAsync(false);
        _categoryRepo.Setup(r => r.SaveAsync(It.IsAny<Category>(), true, default)).Returns(Task.CompletedTask);

        var handler = new CreateCategoryHandler(_uow, _cache.Object, _events.Object, NullLogger<CreateCategoryHandler>.Instance);
        var result = await handler.HandleAsync(new CreateCategoryDto { Name = "Test", Slug = "cat-slug" });

        result.Should().NotBeNull();
        result.Name.Should().Be("Test");
        result.ParentId.Should().BeNull();
        result.Depth.Should().Be(0);
        result.Path.Should().BeEmpty();
    }

    [Fact]
    public async Task CreateCategory_DuplicateSlug_ThrowsConflict()
    {
        _categoryRepo.Setup(r => r.ExistsBySlugAsync("dup-slug", default)).ReturnsAsync(true);
        var handler = new CreateCategoryHandler(_uow, _cache.Object, _events.Object, NullLogger<CreateCategoryHandler>.Instance);
        var act = async () => await handler.HandleAsync(new CreateCategoryDto { Name = "T", Slug = "dup-slug" });
        await act.Should().ThrowAsync<ConflictException>();
    }

    [Fact]
    public async Task CreateCategory_InvalidParent_ThrowsNotFoundException()
    {
        _categoryRepo.Setup(r => r.ExistsBySlugAsync("slug", default)).ReturnsAsync(false);
        _categoryRepo.Setup(r => r.GetByIdAsync("bad-parent", default)).ReturnsAsync((Category?)null);
        var handler = new CreateCategoryHandler(_uow, _cache.Object, _events.Object, NullLogger<CreateCategoryHandler>.Instance);
        var act = async () => await handler.HandleAsync(new CreateCategoryDto { Name = "T", Slug = "slug", ParentId = "bad-parent" });
        await act.Should().ThrowAsync<NotFoundException>();
    }

    [Fact]
    public async Task CreateCategory_WithValidParent_SetsPathAndDepth()
    {
        var parent = new Category { Name = "Parent", Slug = "parent", Path = ["root-id"], Depth = 1 };
        _categoryRepo.Setup(r => r.ExistsBySlugAsync("child-slug", default)).ReturnsAsync(false);
        _categoryRepo.Setup(r => r.GetByIdAsync(parent.Id, default)).ReturnsAsync(parent);
        _categoryRepo.Setup(r => r.SaveAsync(It.IsAny<Category>(), true, default)).Returns(Task.CompletedTask);

        var handler = new CreateCategoryHandler(_uow, _cache.Object, _events.Object, NullLogger<CreateCategoryHandler>.Instance);
        var result = await handler.HandleAsync(new CreateCategoryDto { Name = "Child", Slug = "child-slug", ParentId = parent.Id });

        result.ParentId.Should().Be(parent.Id);
        result.Depth.Should().Be(2);
        result.Path.Should().Contain("root-id");
        result.Path.Should().Contain(parent.Id);
    }

    [Fact]
    public async Task UpdateCategory_CircularParent_ThrowsDomainException()
    {
        var cat2 = new Category { Name = "Cat2", Slug = "cat2" };
        var parentWithCircle = new Category { Name = "P", Slug = "p" };
        // parentWithCircle.Path will contain cat2.Id to create circular reference
        parentWithCircle.Path = [cat2.Id];

        _categoryRepo.Setup(r => r.GetByIdAsync(cat2.Id, default)).ReturnsAsync(cat2);
        _categoryRepo.Setup(r => r.GetByIdAsync(parentWithCircle.Id, default)).ReturnsAsync(parentWithCircle);

        var handler = new UpdateCategoryHandler(_uow, _cache.Object, NullLogger<UpdateCategoryHandler>.Instance);
        var act = async () => await handler.HandleAsync(cat2.Id, new UpdateCategoryDto { Name = "Cat2", ParentId = parentWithCircle.Id, IsActive = true });
        var exception = await act.Should().ThrowAsync<DomainException>();
        exception.WithMessage("*circular*");
    }

    [Fact]
    public async Task UpdateCategory_Success_ReturnsUpdatedDto()
    {
        var category = new Category { Name = "Old", Slug = "old" };
        _categoryRepo.Setup(r => r.GetByIdAsync(category.Id, default)).ReturnsAsync(category);
        _categoryRepo.Setup(r => r.UpdateAsync(It.IsAny<Category>(), default)).Returns(Task.CompletedTask);
        _categoryRepo.Setup(r => r.GetChildrenAsync(It.IsAny<string>(), default)).ReturnsAsync(new List<Category>());

        var handler = new UpdateCategoryHandler(_uow, _cache.Object, NullLogger<UpdateCategoryHandler>.Instance);
        var result = await handler.HandleAsync(category.Id, new UpdateCategoryDto { Name = "New", IsActive = true });

        result.Name.Should().Be("New");
    }

    [Fact]
    public async Task UpdateCategory_NotFound_ThrowsNotFoundException()
    {
        _categoryRepo.Setup(r => r.GetByIdAsync("missing", default)).ReturnsAsync((Category?)null);
        var handler = new UpdateCategoryHandler(_uow, _cache.Object, NullLogger<UpdateCategoryHandler>.Instance);
        var act = async () => await handler.HandleAsync("missing", new UpdateCategoryDto { Name = "X", IsActive = true });
        await act.Should().ThrowAsync<NotFoundException>();
    }

    [Fact]
    public async Task DeleteCategory_HasProducts_ThrowsConflict()
    {
        var category = new Category { Name = "C", Slug = "c" };
        _categoryRepo.Setup(r => r.GetByIdAsync(category.Id, default)).ReturnsAsync(category);
        _categoryRepo.Setup(r => r.HasProductsAsync(category.Id, default)).ReturnsAsync(true);

        var handler = new DeleteCategoryHandler(_uow, _cache.Object, NullLogger<DeleteCategoryHandler>.Instance);
        var act = async () => await handler.HandleAsync(category.Id);
        var exception = await act.Should().ThrowAsync<ConflictException>();
        exception.WithMessage("*products*");
    }

    [Fact]
    public async Task DeleteCategory_HasChildren_ThrowsConflict()
    {
        var category = new Category { Name = "C", Slug = "c" };
        _categoryRepo.Setup(r => r.GetByIdAsync(category.Id, default)).ReturnsAsync(category);
        _categoryRepo.Setup(r => r.HasProductsAsync(category.Id, default)).ReturnsAsync(false);
        _categoryRepo.Setup(r => r.HasChildrenAsync(category.Id, default)).ReturnsAsync(true);

        var handler = new DeleteCategoryHandler(_uow, _cache.Object, NullLogger<DeleteCategoryHandler>.Instance);
        var act = async () => await handler.HandleAsync(category.Id);
        var exception = await act.Should().ThrowAsync<ConflictException>();
        exception.WithMessage("*child*");
    }

    [Fact]
    public async Task DeleteCategory_Success_Deletes()
    {
        var category = new Category { Name = "C", Slug = "c" };
        _categoryRepo.Setup(r => r.GetByIdAsync(category.Id, default)).ReturnsAsync(category);
        _categoryRepo.Setup(r => r.HasProductsAsync(category.Id, default)).ReturnsAsync(false);
        _categoryRepo.Setup(r => r.HasChildrenAsync(category.Id, default)).ReturnsAsync(false);
        _categoryRepo.Setup(r => r.DeleteAsync(category.Id, default)).Returns(Task.CompletedTask);

        var handler = new DeleteCategoryHandler(_uow, _cache.Object, NullLogger<DeleteCategoryHandler>.Instance);
        var act = async () => await handler.HandleAsync(category.Id);
        await act.Should().NotThrowAsync();
        _categoryRepo.Verify(r => r.DeleteAsync(category.Id, default), Times.Once);
    }

    [Fact]
    public async Task DeleteCategory_NotFound_ThrowsNotFoundException()
    {
        _categoryRepo.Setup(r => r.GetByIdAsync("missing", default)).ReturnsAsync((Category?)null);
        var handler = new DeleteCategoryHandler(_uow, _cache.Object, NullLogger<DeleteCategoryHandler>.Instance);
        var act = async () => await handler.HandleAsync("missing");
        await act.Should().ThrowAsync<NotFoundException>();
    }

    [Fact]
    public async Task GetCategoryTree_ReturnsNestedStructure()
    {
        var root = new Category { Name = "Root", Slug = "root" };
        var child = new Category { Name = "Child", Slug = "child", ParentId = root.Id, Depth = 1 };
        _categoryRepo.Setup(r => r.GetAllActiveAsync(default)).ReturnsAsync(new List<Category> { root, child });

        var handler = new GetCategoryTreeHandler(_uow, _cache.Object, NullLogger<GetCategoryTreeHandler>.Instance);
        var result = await handler.HandleAsync();

        result.Should().HaveCount(1);
        result[0].Children.Should().HaveCount(1);
        result[0].Children[0].Name.Should().Be("Child");
    }

    [Fact]
    public async Task GetCategoryTree_Empty_ReturnsEmptyArray()
    {
        _categoryRepo.Setup(r => r.GetAllActiveAsync(default)).ReturnsAsync(new List<Category>());

        var handler = new GetCategoryTreeHandler(_uow, _cache.Object, NullLogger<GetCategoryTreeHandler>.Instance);
        var result = await handler.HandleAsync();

        result.Should().BeEmpty();
    }
}

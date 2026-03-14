using Catalog.Api.Domain.Entities;
using Catalog.Api.Domain.Interfaces;
using MongoDB.Driver;

namespace Catalog.Api.Infrastructure.Repositories;

/// <summary>
/// MongoDB-backed repository for <see cref="Category"/> persistence.
/// </summary>
public sealed class CategoryRepository(IMongoDatabase database, ILogger<CategoryRepository> logger)
    : MongoRepositoryBase<Category>(logger), ICategoryRepository
{
    private readonly IMongoCollection<Category> _categories = database.GetCollection<Category>("categories");
    private readonly IMongoCollection<Product> _products = database.GetCollection<Product>("products");

    /// <inheritdoc />
    public Task<Category?> GetByIdAsync(string id, CancellationToken cancellationToken = default) =>
        ExecuteAsync<Category?>(
            () => _categories.Find(c => c.Id == id).FirstOrDefaultAsync(cancellationToken)!,
            "Error retrieving category {CategoryId}", id);

    /// <inheritdoc />
    public Task CreateAsync(Category category, CancellationToken cancellationToken = default) =>
        ExecuteAsync(
            () => _categories.InsertOneAsync(category, cancellationToken: cancellationToken),
            "Error creating category {CategoryId}", category.Id);

    /// <inheritdoc />
    public Task UpdateAsync(Category category, CancellationToken cancellationToken = default) =>
        ExecuteAsync(
            () => _categories.ReplaceOneAsync(
                existing => existing.Id == category.Id,
                category,
                cancellationToken: cancellationToken),
            "Error updating category {CategoryId}", category.Id);

    /// <inheritdoc />
    public Task<bool> ExistsBySlugAsync(string slug, CancellationToken cancellationToken = default) =>
        ExecuteAsync(
            () => _categories.Find(c => c.Slug == slug).AnyAsync(cancellationToken),
            "Error checking slug existence {Slug}", slug);

    /// <inheritdoc />
    public Task<Category?> GetBySlugAsync(string slug, CancellationToken cancellationToken = default) =>
        ExecuteAsync<Category?>(
            () => _categories.Find(c => c.Slug == slug).FirstOrDefaultAsync(cancellationToken)!,
            "Error retrieving category by slug {Slug}", slug);

    /// <inheritdoc />
    public Task<IReadOnlyList<Category>> GetAllActiveAsync(CancellationToken cancellationToken = default) =>
        ExecuteAsync<IReadOnlyList<Category>>(
            async () => await _categories.Find(c => c.IsActive).ToListAsync(cancellationToken),
            "Error retrieving all active categories");

    /// <inheritdoc />
    public Task<IReadOnlyList<Category>> GetChildrenAsync(string parentId, CancellationToken cancellationToken = default) =>
        ExecuteAsync<IReadOnlyList<Category>>(
            async () => await _categories.Find(c => c.ParentId == parentId).ToListAsync(cancellationToken),
            "Error retrieving children for category {CategoryId}", parentId);

    /// <inheritdoc />
    public Task<bool> HasProductsAsync(string categoryId, CancellationToken cancellationToken = default) =>
        ExecuteAsync(
            () => _products.Find(p => p.CategoryIds.Contains(categoryId)).AnyAsync(cancellationToken),
            "Error checking product usage for category {CategoryId}", categoryId);

    /// <inheritdoc />
    public Task<bool> HasChildrenAsync(string categoryId, CancellationToken cancellationToken = default) =>
        ExecuteAsync(
            () => _categories.Find(c => c.ParentId == categoryId).AnyAsync(cancellationToken),
            "Error checking children for category {CategoryId}", categoryId);

    /// <inheritdoc />
    public Task<IReadOnlyList<string>> GetDescendantIdsAsync(string categoryId, CancellationToken cancellationToken = default) =>
        ExecuteAsync<IReadOnlyList<string>>(
            async () =>
            {
                var filter = Builders<Category>.Filter.AnyEq(c => c.Path, categoryId);
                var descendants = await _categories.Find(filter).ToListAsync(cancellationToken);
                return descendants.Select(c => c.Id).ToList();
            },
            "Error retrieving descendant IDs for category {CategoryId}", categoryId);

    /// <inheritdoc />
    public Task DeleteAsync(string id, CancellationToken cancellationToken = default) =>
        ExecuteAsync(
            () => _categories.DeleteOneAsync(c => c.Id == id, cancellationToken),
            "Error deleting category {CategoryId}", id);
}

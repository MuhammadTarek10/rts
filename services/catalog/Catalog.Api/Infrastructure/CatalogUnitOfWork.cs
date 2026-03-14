using Catalog.Api.Domain.Interfaces;

namespace Catalog.Api.Infrastructure;

public sealed class CatalogUnitOfWork(
    IProductRepository products,
    ICategoryRepository categories,
    IBrandRepository brands)
{
    public IProductRepository Products { get; } = products;
    public ICategoryRepository Categories { get; } = categories;
    public IBrandRepository Brands { get; } = brands;
}

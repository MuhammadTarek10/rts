using Catalog.Api.Shared.Abstraction;

namespace Catalog.Api.Infrastructure.Messaging;

public interface ICatalogEventPublisher
{
    Task PublishAsync(IDomainEvent domainEvent, CancellationToken cancellationToken = default);
}

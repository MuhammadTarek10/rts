namespace Catalog.Api.Shared.Exceptions;

public sealed class NotFoundException(string code, string message) : DomainException(code, message);

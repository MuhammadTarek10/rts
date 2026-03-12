namespace Catalog.Api.Shared.Exceptions;

/// <summary>
/// Thrown when an operation conflicts with the current state of a resource (e.g. duplicate SKU).
/// </summary>
public sealed class ConflictException(string code, string message) : DomainException(code, message);

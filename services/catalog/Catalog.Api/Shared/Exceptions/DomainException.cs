namespace Catalog.Api.Shared.Exceptions;

/// <summary>
/// Base exception for domain-level errors that carry a machine-readable code.
/// </summary>
public class DomainException(string code, string message) : Exception(message)
{
    /// <summary>
    /// Machine-readable error code identifying the type of domain violation.
    /// </summary>
    public string Code { get; } = code;
}

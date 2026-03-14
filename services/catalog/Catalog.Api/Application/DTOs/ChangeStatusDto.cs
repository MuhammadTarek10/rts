using System.ComponentModel.DataAnnotations;

namespace Catalog.Api.Application.DTOs;

public sealed class ChangeStatusDto
{
    [Required] public string Status { get; init; } = string.Empty;
}

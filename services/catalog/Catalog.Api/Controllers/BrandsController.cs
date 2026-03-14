using Catalog.Api.Application.DTOs;
using Catalog.Api.Application.Handlers;
using Catalog.Api.Infrastructure;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;

namespace Catalog.Api.Controllers;

[ApiController]
[Route("api/brands")]
public sealed class BrandsController(
    CreateBrandHandler createBrandHandler,
    UpdateBrandHandler updateBrandHandler,
    DeleteBrandHandler deleteBrandHandler,
    CatalogUnitOfWork uow) : ControllerBase
{
    [HttpPost]
    [Authorize(Policy = "Admin")]
    [ProducesResponseType(typeof(BrandResponseDto), StatusCodes.Status201Created)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    [ProducesResponseType(StatusCodes.Status401Unauthorized)]
    [ProducesResponseType(StatusCodes.Status403Forbidden)]
    [ProducesResponseType(StatusCodes.Status409Conflict)]
    public async Task<IActionResult> Create([FromBody] CreateBrandDto request, CancellationToken cancellationToken)
    {
        var brand = await createBrandHandler.HandleAsync(request, cancellationToken);
        return CreatedAtAction(nameof(GetById), new { id = brand.Id }, brand);
    }

    [HttpGet("{id}")]
    [AllowAnonymous]
    [ProducesResponseType(typeof(BrandResponseDto), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> GetById([FromRoute] string id, CancellationToken cancellationToken)
    {
        var brand = await uow.Brands.GetByIdAsync(id, cancellationToken);
        if (brand is null) return NotFound();
        return Ok(BrandResponseDto.FromEntity(brand));
    }

    [HttpGet]
    [AllowAnonymous]
    [ProducesResponseType(typeof(IReadOnlyList<BrandResponseDto>), StatusCodes.Status200OK)]
    public async Task<IActionResult> GetAll(CancellationToken cancellationToken)
    {
        var brands = await uow.Brands.GetAllActiveAsync(cancellationToken);
        return Ok(brands.Select(BrandResponseDto.FromEntity).ToList());
    }

    [HttpPut("{id}")]
    [Authorize(Policy = "Admin")]
    [ProducesResponseType(typeof(BrandResponseDto), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    [ProducesResponseType(StatusCodes.Status401Unauthorized)]
    [ProducesResponseType(StatusCodes.Status403Forbidden)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> Update([FromRoute] string id, [FromBody] UpdateBrandDto request, CancellationToken cancellationToken)
    {
        var brand = await updateBrandHandler.HandleAsync(id, request, cancellationToken);
        return Ok(brand);
    }

    [HttpDelete("{id}")]
    [Authorize(Policy = "Admin")]
    [ProducesResponseType(StatusCodes.Status204NoContent)]
    [ProducesResponseType(StatusCodes.Status401Unauthorized)]
    [ProducesResponseType(StatusCodes.Status403Forbidden)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    [ProducesResponseType(StatusCodes.Status409Conflict)]
    public async Task<IActionResult> Delete([FromRoute] string id, CancellationToken cancellationToken)
    {
        await deleteBrandHandler.HandleAsync(id, cancellationToken);
        return NoContent();
    }
}

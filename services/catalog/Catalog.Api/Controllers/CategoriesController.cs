using Catalog.Api.Application.DTOs;
using Catalog.Api.Application.Handlers;
using Catalog.Api.Infrastructure;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;

namespace Catalog.Api.Controllers;

[ApiController]
[Route("api/categories")]
public sealed class CategoriesController(
    CreateCategoryHandler createCategoryHandler,
    UpdateCategoryHandler updateCategoryHandler,
    DeleteCategoryHandler deleteCategoryHandler,
    GetCategoryTreeHandler getCategoryTreeHandler) : ControllerBase
{
    [HttpPost]
    [Authorize(Policy = "Admin")]
    [ProducesResponseType(typeof(CategoryResponseDto), StatusCodes.Status201Created)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    [ProducesResponseType(StatusCodes.Status401Unauthorized)]
    [ProducesResponseType(StatusCodes.Status403Forbidden)]
    [ProducesResponseType(StatusCodes.Status409Conflict)]
    public async Task<IActionResult> Create([FromBody] CreateCategoryDto request, CancellationToken cancellationToken)
    {
        var category = await createCategoryHandler.HandleAsync(request, cancellationToken);
        return CreatedAtAction(nameof(GetById), new { id = category.Id }, category);
    }

    [HttpGet]
    [AllowAnonymous]
    [ProducesResponseType(typeof(IReadOnlyList<CategoryResponseDto>), StatusCodes.Status200OK)]
    public async Task<IActionResult> GetAll(CancellationToken cancellationToken)
    {
        var tree = await getCategoryTreeHandler.HandleAsync(cancellationToken);
        return Ok(tree);
    }

    [HttpGet("tree")]
    [AllowAnonymous]
    [ProducesResponseType(typeof(IReadOnlyList<CategoryTreeDto>), StatusCodes.Status200OK)]
    public async Task<IActionResult> GetTree(CancellationToken cancellationToken)
    {
        var tree = await getCategoryTreeHandler.HandleAsync(cancellationToken);
        return Ok(tree);
    }

    [HttpGet("{id}/children")]
    [AllowAnonymous]
    [ProducesResponseType(typeof(IReadOnlyList<CategoryResponseDto>), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> GetChildren([FromRoute] string id, [FromServices] CatalogUnitOfWork uow, CancellationToken cancellationToken)
    {
        var children = await uow.Categories.GetChildrenAsync(id, cancellationToken);
        return Ok(children.Select(CategoryResponseDto.FromEntity).ToList());
    }

    [HttpGet("{id}")]
    [AllowAnonymous]
    [ProducesResponseType(typeof(CategoryResponseDto), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> GetById([FromRoute] string id, [FromServices] CatalogUnitOfWork uow, CancellationToken cancellationToken)
    {
        var category = await uow.Categories.GetByIdAsync(id, cancellationToken);
        if (category is null) return NotFound();
        return Ok(CategoryResponseDto.FromEntity(category));
    }

    [HttpPut("{id}")]
    [Authorize(Policy = "Admin")]
    [ProducesResponseType(typeof(CategoryResponseDto), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    [ProducesResponseType(StatusCodes.Status401Unauthorized)]
    [ProducesResponseType(StatusCodes.Status403Forbidden)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> Update([FromRoute] string id, [FromBody] UpdateCategoryDto request, CancellationToken cancellationToken)
    {
        var category = await updateCategoryHandler.HandleAsync(id, request, cancellationToken);
        return Ok(category);
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
        await deleteCategoryHandler.HandleAsync(id, cancellationToken);
        return NoContent();
    }
}

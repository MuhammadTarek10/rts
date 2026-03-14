using Catalog.Api.Application.DTOs;
using Catalog.Api.Application.Handlers;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;

namespace Catalog.Api.Controllers;

[ApiController]
[Route("api/products")]
public sealed class ProductsController(
    CreateProductHandler createProductHandler,
    UpdateProductHandler updateProductHandler,
    DeleteProductHandler deleteProductHandler,
    ChangeProductStatusHandler changeProductStatusHandler,
    SearchProductsHandler searchProductsHandler,
    GetProductByIdHandler getProductByIdHandler,
    GetProductBySlugHandler getProductBySlugHandler,
    UploadProductImageHandler uploadProductImageHandler,
    DeleteProductImageHandler deleteProductImageHandler,
    ReorderProductImagesHandler reorderProductImagesHandler) : ControllerBase
{
    [HttpPost]
    [Authorize(Policy = "Admin")]
    [ProducesResponseType(typeof(ProductResponseDto), StatusCodes.Status201Created)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    [ProducesResponseType(StatusCodes.Status401Unauthorized)]
    [ProducesResponseType(StatusCodes.Status403Forbidden)]
    [ProducesResponseType(StatusCodes.Status409Conflict)]
    public async Task<IActionResult> Create([FromBody] CreateProductDto request, CancellationToken cancellationToken)
    {
        var product = await createProductHandler.HandleAsync(request, cancellationToken);
        return CreatedAtAction(nameof(GetById), new { id = product.Id }, product);
    }

    [HttpGet("{id}")]
    [AllowAnonymous]
    [ProducesResponseType(typeof(ProductResponseDto), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> GetById([FromRoute] string id, CancellationToken cancellationToken)
    {
        var product = await getProductByIdHandler.HandleAsync(id, cancellationToken);
        return Ok(product);
    }

    [HttpGet("slug/{slug}")]
    [AllowAnonymous]
    [ProducesResponseType(typeof(ProductResponseDto), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> GetBySlug([FromRoute] string slug, CancellationToken cancellationToken)
    {
        var product = await getProductBySlugHandler.HandleAsync(slug, cancellationToken);
        return Ok(product);
    }

    [HttpGet]
    [AllowAnonymous]
    [ProducesResponseType(typeof(SearchProductsResponse), StatusCodes.Status200OK)]
    public async Task<IActionResult> Search([FromQuery] SearchProductsRequest request, CancellationToken cancellationToken)
    {
        var result = await searchProductsHandler.HandleAsync(request, cancellationToken);
        return Ok(result);
    }

    [HttpPut("{id}")]
    [Authorize(Policy = "Admin")]
    [ProducesResponseType(typeof(ProductResponseDto), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    [ProducesResponseType(StatusCodes.Status401Unauthorized)]
    [ProducesResponseType(StatusCodes.Status403Forbidden)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    [ProducesResponseType(StatusCodes.Status409Conflict)]
    public async Task<IActionResult> Update([FromRoute] string id, [FromBody] UpdateProductDto request, CancellationToken cancellationToken)
    {
        var product = await updateProductHandler.HandleAsync(id, request, cancellationToken);
        return Ok(product);
    }

    [HttpDelete("{id}")]
    [Authorize(Policy = "Admin")]
    [ProducesResponseType(StatusCodes.Status204NoContent)]
    [ProducesResponseType(StatusCodes.Status401Unauthorized)]
    [ProducesResponseType(StatusCodes.Status403Forbidden)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> Delete([FromRoute] string id, CancellationToken cancellationToken)
    {
        await deleteProductHandler.HandleAsync(id, cancellationToken);
        return NoContent();
    }

    [HttpPatch("{id}/status")]
    [Authorize(Policy = "Admin")]
    [ProducesResponseType(typeof(ProductResponseDto), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    [ProducesResponseType(StatusCodes.Status401Unauthorized)]
    [ProducesResponseType(StatusCodes.Status403Forbidden)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> ChangeStatus([FromRoute] string id, [FromBody] ChangeStatusDto request, CancellationToken cancellationToken)
    {
        var product = await changeProductStatusHandler.HandleAsync(id, request, cancellationToken);
        return Ok(product);
    }

    [HttpPost("{id}/images")]
    [Authorize(Policy = "Admin")]
    [ProducesResponseType(typeof(ProductResponseDto), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    [ProducesResponseType(StatusCodes.Status401Unauthorized)]
    [ProducesResponseType(StatusCodes.Status403Forbidden)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> UploadImage(
        [FromRoute] string id,
        IFormFile file,
        [FromForm] string? altText,
        CancellationToken cancellationToken)
    {
        var product = await uploadProductImageHandler.HandleAsync(id, file, altText, cancellationToken);
        return Ok(product);
    }

    [HttpDelete("{id}/images/{imageId}")]
    [Authorize(Policy = "Admin")]
    [ProducesResponseType(typeof(ProductResponseDto), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status401Unauthorized)]
    [ProducesResponseType(StatusCodes.Status403Forbidden)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> DeleteImage([FromRoute] string id, [FromRoute] string imageId, CancellationToken cancellationToken)
    {
        var product = await deleteProductImageHandler.HandleAsync(id, imageId, cancellationToken);
        return Ok(product);
    }

    [HttpPut("{id}/images/order")]
    [Authorize(Policy = "Admin")]
    [ProducesResponseType(typeof(ProductResponseDto), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    [ProducesResponseType(StatusCodes.Status401Unauthorized)]
    [ProducesResponseType(StatusCodes.Status403Forbidden)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> ReorderImages([FromRoute] string id, [FromBody] ReorderImagesDto request, CancellationToken cancellationToken)
    {
        var product = await reorderProductImagesHandler.HandleAsync(id, request, cancellationToken);
        return Ok(product);
    }
}

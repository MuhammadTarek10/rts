using Catalog.Api.Application.DTOs;
using Catalog.Api.Application.Handlers;
using Catalog.Api.Domain.Interfaces;
using Microsoft.AspNetCore.Mvc;

namespace Catalog.Api.Controllers;

/// <summary>
/// API controller for product catalog operations.
/// </summary>
[ApiController]
[Route("api/products")]
public sealed class ProductsController(CreateProductHandler createProductHandler, IProductRepository productRepository) : ControllerBase
{
    /// <summary>
    /// Creates a new product in the catalog.
    /// </summary>
    /// <param name="request">Product creation payload.</param>
    /// <param name="cancellationToken">Cancellation token.</param>
    /// <returns>The newly created product.</returns>
    [HttpPost]
    [ProducesResponseType(typeof(ProductResponseDto), StatusCodes.Status201Created)]
    [ProducesResponseType(StatusCodes.Status400BadRequest)]
    [ProducesResponseType(StatusCodes.Status409Conflict)]
    public async Task<IActionResult> Create([FromBody] CreateProductDto request, CancellationToken cancellationToken)
    {
        var product = await createProductHandler.HandleAsync(request, cancellationToken);

        return CreatedAtAction(nameof(GetById), new { id = product.Id }, product);
    }

    /// <summary>
    /// Retrieves a single product by its identifier.
    /// </summary>
    /// <param name="id">The product identifier.</param>
    /// <param name="cancellationToken">Cancellation token.</param>
    /// <returns>The product if found; otherwise 404.</returns>
    [HttpGet("{id}")]
    [ProducesResponseType(typeof(ProductResponseDto), StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound)]
    public async Task<IActionResult> GetById([FromRoute] string id, CancellationToken cancellationToken)
    {
        var product = await productRepository.GetByIdAsync(id, cancellationToken);

        if (product is null)
        {
            return NotFound();
        }

        return Ok(CreateProductHandler.Map(product));
    }
}

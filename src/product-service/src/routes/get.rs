use crate::model::ProductInfo;
use crate::startup::AppState;
use actix_web::{web, Error, HttpResponse};
use tracing::{info, instrument, warn};

#[instrument(name = "get_product", skip(data), fields(product_id = %path.product_id))]
pub async fn get_product(
    data: web::Data<AppState>,
    path: web::Path<ProductInfo>,
) -> Result<HttpResponse, Error> {
    let products = data.products.lock().unwrap();

    // find product by id in products
    let index = products.iter().position(|p| p.id == path.product_id);
    if let Some(i) = index {
        info!(product_name = %products[i].name, "Product found");
        return Ok(HttpResponse::Ok().json(products[i].clone()));
    } else {
        warn!(product_id = %path.product_id, "Product not found");
        return Ok(HttpResponse::NotFound().body("Product not found"));
    }
}

#[instrument(name = "get_all_products", skip(data))]
pub async fn get_products(data: web::Data<AppState>) -> Result<HttpResponse, Error> {
    let products = data.products.lock().unwrap();
    let count = products.len();
    info!(product_count = count, "Retrieved all products");
    Ok(HttpResponse::Ok().json(products.to_vec()))
}

pub async fn get_metrics(data: web::Data<AppState>) -> Result<HttpResponse, Error> {
    let products = data.products.lock().unwrap();
    let total_product_count = products.len();
    let metrics = "# HELP total_product_count The total number of products in the database\n# TYPE total_product_count gauge\ntotal_product_count ".to_string() + &total_product_count.to_string() + "\n";
    Ok(HttpResponse::Ok().body(metrics))
}

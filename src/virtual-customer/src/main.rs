use rand::Rng;
use serde::Serialize;
use std::env;
use std::thread;
use std::time::{Duration, Instant};
use tracing::{info, instrument, warn, error};
mod telemetry;

#[instrument]
fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Initialize OpenTelemetry
    if let Err(e) = telemetry::init_tracer() {
        eprintln!("Failed to initialize OpenTelemetry tracer: {}", e);
    }
    
    let order_service_url =
        env::var("ORDER_SERVICE_URL").unwrap_or_else(|_| "http://localhost:3000".to_string());
    info!(order_service_url = %order_service_url, "Using order service");

    let orders_per_hour: u64 = env::var("ORDERS_PER_HOUR")
        .unwrap_or_else(|_| "1".to_string())
        .parse()
        .unwrap_or(1);
    info!(orders_per_hour = orders_per_hour, "Configured orders per hour");

    if orders_per_hour == 0 {
        println!("Please set the ORDERS_PER_HOUR environment variable");
        std::process::exit(1);
    }

    println!("Orders to submit per hour: {}", orders_per_hour);

    let minutes: f64 = 60.0;
    let seconds: f64 = 60.0;

    // calculate the time between orders in seconds
    let order_submission_interval: f64 = (minutes / (orders_per_hour as f64)) * seconds;
    println!(
        "Order submission interval: {} seconds",
        order_submission_interval
    );

    // time to sleep between orders
    let sleep_duration = Duration::from_secs_f64(order_submission_interval);
    println!("Sleep duration between orders: {:?}", sleep_duration);

    // order counter
    let mut order_counter = 0;

    // keep track of how long we've been running
    let start_time = Instant::now();

    loop {
        // Create a span for this order processing cycle
        let _span = tracing::span!(tracing::Level::INFO, "process_order").entered();
        
        order_counter += 1;

        // generate a random customer id
        let customer_id = (rand::thread_rng().gen_range(1000000000..2147483647)).to_string();
        
        // generate a random number of items to order
        let number_of_items = rand::thread_rng().gen_range(1..5);
        info!(customer_id = %customer_id, items_count = number_of_items, "Creating new order");

        // create a vector to hold the items
        let mut items = Vec::new();

        // generate a random item for each item
        items.append(
            &mut (0..number_of_items)
                .map(|_| {
                    let product_id = rand::thread_rng().gen_range(1..10);
                    let quantity = rand::thread_rng().gen_range(1..5);
                    let price = rand::thread_rng().gen_range(1.0..100.0);

                    Item {
                        product_id,
                        quantity,
                        price,
                    }
                })
                .collect(),
        );

        let order = Order { customer_id, items };
        let serialized_order = serde_json::to_string(&order)?;
        
        // Create a span for the HTTP request
        let _span = tracing::span!(tracing::Level::INFO, "submit_order", 
            customer_id = %order.customer_id, 
            items_count = order.items.len()
        ).entered();
        
        let client = reqwest::blocking::Client::new();
        info!(url = %order_service_url, "Sending order request");
        
        let response = client
            .post(order_service_url.clone())
            .header("Content-Type", "application/json")
            .body(serialized_order.clone())
            .send();

        match response {
            Ok(res) => {
                // Handle successful response
                let elapsed_time = start_time.elapsed();
                let status = res.status();
                
                // Log success with tracing
                info!(
                    order_number = order_counter,
                    elapsed_time = ?elapsed_time,
                    status = %status,
                    "Order successfully submitted"
                );

                // print the order details
                println!(
                    "Order {} sent at {:.2?} with status of {}. {}",
                    order_counter,
                    elapsed_time,
                    status,
                    serialized_order
                );
            }
            Err(err) => {
                // Log error with tracing
                error!(
                    order_number = order_counter,
                    error = %err,
                    "Failed to submit order"
                );
                
                // Handle error
                println!("Failed to submit order: {}", err);
            }
        }

        thread::sleep(sleep_duration);
    }
    
    // Shutdown the tracer provider before the program exits
    telemetry::shutdown_tracer();
    Ok(())
}

#[derive(Serialize, Debug)]
struct Order {
    #[serde(rename = "customerId")]
    customer_id: String,
    items: Vec<Item>,
}

#[derive(Serialize, Debug)]
struct Item {
    #[serde(rename = "productId")]
    product_id: u32,
    quantity: u32,
    price: f32,
}

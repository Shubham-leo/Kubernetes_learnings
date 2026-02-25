use actix_web::{web, App, HttpServer, HttpResponse};
use serde_json::json;
use std::time::Instant;

fn fib(n: u64) -> u64 {
    if n <= 1 { return n; }
    fib(n - 1) + fib(n - 2)
}

async fn index() -> HttpResponse {
    let start = Instant::now();
    let result = fib(30);
    let duration = start.elapsed().as_millis();

    HttpResponse::Ok().json(json!({
        "language": "Rust",
        "fibonacci_30": result,
        "compute_time_ms": duration,
    }))
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    println!("Rust server on :8090");
    HttpServer::new(|| {
        App::new().route("/", web::get().to(index))
    })
    .bind("0.0.0.0:8090")?
    .run()
    .await
}

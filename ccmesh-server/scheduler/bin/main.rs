use hyper::service::{make_service_fn, service_fn};
use hyper::{Body, Request, Response, Server};
use hz_config::*;
use scheduler::mesh::{mesh_service_x, mesh_service0, mesh_service2, mesh_service3};
use scheduler::mesg::{mesg_service};
use scheduler::cb::{cb_service, cb_service0, cb_service2, cb_service3};
use std::convert::Infallible;
use tracing::info;

async fn service0(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
    match MODE {
        "mesh" => mesh_service0(_req).await,
        "cb" => cb_service0(_req).await,
        _ => panic!("unknown mode"),
    }
}

async fn service(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
    match MODE {
        "mesh" => mesh_service_x(_req).await,
        "cb" => cb_service(_req).await,
        "ccmesg" => mesg_service(_req).await,
        _ => panic!("unknown mode"),
    }
}

async fn service2(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
    match MODE {
        "mesh" => {
            // mesh_service2(_req).await
            if rand::random() {
                mesh_service2(_req).await
            } else {
                mesh_service3(_req).await
            }
        }
        "cb" => {
            // mesh_service2(_req).await
            if rand::random() {
                cb_service2(_req).await
            } else {
                cb_service3(_req).await
            }
        }
        _ => panic!("unknown mode"),
    }
}

#[tokio::main(worker_threads = 6)]
async fn main() {
    tracing_subscriber::fmt::init();
    info!("Current Mode: {:?}", MODE);
    match MODE {
        "mesh" => {}
        "cb" => hz_cb::goclient::setup_clients(),
        "ccmesg" => {},
        _ => panic!("unknown mode or not implemented"),
    }
    let addr = std::net::SocketAddr::from(([127, 0, 0, 1], 3000));
    let make_svc = make_service_fn(|_conn| async { Ok::<_, Infallible>(service_fn(service)) });
    let server = Server::bind(&addr).serve(make_svc);
    server.await.unwrap();
}

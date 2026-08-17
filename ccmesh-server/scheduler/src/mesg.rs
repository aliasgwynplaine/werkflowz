use crate::utils::*;
use hyper::{Body, Request, Response};
use hz_config::*;
use ccmesg::goclient::Envelope;
use std::convert::Infallible;


pub async fn mesg_service(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
    let mut e = Envelope::new();
    e.workload = get_2103();
    let req = serde_json::to_string(&e).unwrap();
    let idx = rand::random::<usize>() % T;
    send_req(idx, req, "Entry").await;
    
    Ok(Response::new(Body::from("Ok")))
}
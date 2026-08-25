use crate::utils::*;
use hz_config::*;
use hyper::{Body, Request, Response};
use ccmesgbox::goclient::Envelope;
use std::convert::Infallible;

pub async fn mesgbox_service_long_seq(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
    let mut e = Envelope::new();
    e.workload = get_long_work();
    
    loop {
        let req = serde_json::to_string(&e).unwrap();
        let res = send_req_c(req, "Entry").await;
        let res_e = serde_json::from_slice::<Envelope>(&res).unwrap();
        
        if res_e.abort {
            continue;
        }
        
        break;
    }
    
    Ok(Response::new(Body::from("Ok")))
}

pub async fn mesgbox_service_c(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
    let mut e = Envelope::new();
    e.workload = get_2103();
    
    loop {
        let req = serde_json::to_string(&e).unwrap();
        let res = send_req_c(req, "Entry").await;
        let res_e = serde_json::from_slice::<Envelope>(&res).unwrap();
        
        if res_e.abort {
            continue;
        }
        
        break;
    }
    
    Ok(Response::new(Body::from("Ok")))
}


pub async fn mesgbox_service(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
    let mut e = Envelope::new();
    e.workload = get_2103();
    
    loop {
        let req = serde_json::to_string(&e).unwrap();
        let idx = rand::random::<usize>() % T;
        let res = send_req(idx, req, "Entry").await;
        let res_e = serde_json::from_slice::<Envelope>(&res).unwrap();
        
        if res_e.abort {
            continue;
        }
        
        break;
    }
    
    Ok(Response::new(Body::from("Ok")))
}

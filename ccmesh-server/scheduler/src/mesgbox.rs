use crate::utils::*;
use hz_config::*;
use hyper::{Body, Request, Response};
use ccmesgbox::goclient::Envelope;
use std::convert::Infallible;
use std::net::TcpListener;
use std::io::Read;

pub async fn mesgbox_service_linear_central(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
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

pub async fn mesgbox_service_linear_distrib(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
    let mut e = Envelope::new();
    e.workload = get_long_work();
    
    loop {
        
        let listener = TcpListener::bind("0.0.0.0:0").unwrap();
        let addr = listener.local_addr().unwrap();
        e.port = format!("{}", addr.port());
        
        let req = serde_json::to_string(&e).unwrap();
        let idx = rand::random::<usize>() % T;
        let _res = send_req(idx, req, "Entry").await;

        let (mut res_stream, _) = listener.accept().unwrap();
        let mut buf = [0u8, 4];
        let nb = res_stream.read(&mut buf).unwrap();        

        if nb > 0 {
            let rv = buf[0];
    
            if rv == 0 { // abort
                continue;
            }
        } else {
            continue;
        }
        
        break;
    }
    
    Ok(Response::new(Body::from("Ok")))
}

pub async fn mesgbox_service_fi_fo_central(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
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


pub async fn mesgbox_service_fi_fo_distrib(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
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

use crate::utils::*;
use hz_config::*;
use hyper::{Body, Request, Response};
use ccmesgbox::goclient::Envelope;
use tracing::info;
use std::convert::Infallible;
use std::net::{IpAddr, TcpListener, UdpSocket};
use std::io::{Read};
use std::io;

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

fn get_local_ip() -> io::Result<IpAddr> {
    let socket = UdpSocket::bind("0.0.0.0:0")?;
    // Doesn't need to be reachable; just used for route resolution.
    socket.connect("8.8.8.8:80")?;
    Ok(socket.local_addr()?.ip())
}


pub async fn mesgbox_service_linear_distrib(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
    let mut e = Envelope::new();
    e.workload = get_long_work();
    
    loop {
        
        let listener = TcpListener::bind("0.0.0.0:0").unwrap();
        let addr = listener.local_addr().unwrap();
        e.returnadr = format!("{}:{}", get_local_ip().unwrap(), addr.port());
        
        let req = serde_json::to_string(&e).unwrap();
        let idx = rand::random::<usize>() % T;
        let _res = send_req(idx, req, "Entry").await;

        let (mut res_stream, _) = listener.accept().unwrap();
        let mut buf = [0u8, 8];
        let nb = res_stream.read(&mut buf).unwrap();     
        //info!("r: {:?}", buf[..nb]);   

        if nb > 0 {
    
            match std::str::from_utf8(&buf[..nb]) {
            Ok(text) => {
                let trimmed = text.trim();
                if trimmed == "0" {
                    println!("abort");
                } else {
                    println!("succeed");
                }
            }
            Err(_) => {
                // Not valid UTF-8 -- treat as non-zero/succeed, or handle as you prefer
                println!("succeed");
            }
        }
        } else {
            info!("uu");
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

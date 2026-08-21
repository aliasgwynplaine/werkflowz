use crate::utils::*;
use hyper::{Body, Request, Response};
use ccmesg::goclient::Envelope;
use std::convert::Infallible;


pub async fn mesg_service(_req: Request<Body>) -> Result<Response<Body>, Infallible> {
    let mut e = Envelope::new();
    e.workload = get_2103();
    
    loop {
        let req = serde_json::to_string(&e).unwrap();
        let res = send_req(0, req, "Entry").await;
        let res_e = serde_json::from_slice::<Envelope>(&res).unwrap();
        
        if res_e.abort {
            continue;
        }
        
        break;
    }
    
    Ok(Response::new(Body::from("Ok")))
}

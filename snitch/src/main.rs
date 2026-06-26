use std::collections::HashMap;
use std::sync::{RwLock, Arc};
use std::io::Result;
use tokio::io::{AsyncBufReadExt, AsyncWriteExt, BufReader};
use tokio::net::{TcpListener, TcpStream};

type Tid = u32;
type Fid = String; // not so sure about this
type Addr = String;
type FunctionMap = RwLock<HashMap<Fid, Addr>>; // fid -> "addr:port"
type TransactionMap = Arc<RwLock<HashMap<Tid, Arc<FunctionMap>>>>;

static GATEWAY_ADDR : &str = "localhost:8000";

#[tokio::main]
async fn main() -> Result<()> {
    println!("Hello, friend!");
    let txn : TransactionMap = Arc::new(RwLock::new(HashMap::new()));

    let listener = TcpListener::bind("0.0.0.0:46655").await?;
    println!("Listening in port 46655");

    loop {
        let (socket, addr) = listener.accept().await?;
        let map = Arc::clone(&txn);

        println!("Connection accepted from {addr}");

        tokio::spawn(
            async move {
                if let Err(e) = handle_connection(socket, map).await {
                    eprintln!("error with {addr}: {e}");
                }
            }
        );
    }
}


async fn handle_connection(socket: TcpStream, map: TransactionMap) -> Result<()> {
    let (reader, mut writer) = socket.into_split();
    let mut lines = BufReader::new(reader).lines();

    while let Some(line) = lines.next_line().await? {
        let response = process_line(&line, &map);
        writer.write_all(response.await.as_bytes()).await?;
        writer.write_all(b"\n").await?;
    }

    println!("Closing connection.");

    Ok(())
}


async fn process_line(line: &str, map: &TransactionMap) -> String {
    let parts : Vec<&str> = line.trim().split_whitespace().collect();

    match parts.as_slice() {
        ["GET", tid, fid] => {
            let (tid, fid) = match(tid.parse::<Tid>(), fid.parse::<Fid>()) {
                (Ok(t), Ok(f)) => (t, f),
                _ => return "fuck you".to_string(),
            };
            format!("Ok {}", resolve(map, tid, fid))
        }

        ["PUT", tid, fid, addr] => {
            let tid = match tid.parse::<Tid>() {
                Ok(v) => v, Err(_) => return "tuck you".to_string()
            };
            let fid = match fid.parse::<Fid>() {
                Ok(v) => v, Err(_) => return "fuck you".to_string()
            };
            let addr = match addr.parse::<String>() {
                Ok(v) => v, Err(_) => return "suck you".to_string()
            };

            put(map, tid, fid, addr);

            "Ok".to_string()
        }

        ["COMMIT", tid] => {
            let tid = match tid.parse::<Tid>() {
                Ok(v) => v, Err(_) => return "fuck you".to_string()
            };

            delete(map, tid);

            "Ok".to_string()
        }

        _ => "FUCK YOU".to_string(),
    }
}

fn resolve(map: &TransactionMap, tid: Tid, fid: Fid) -> Addr {
    let transmap = map.read().unwrap();
    match transmap.get(&tid) {
        Some(trans_lock) => {
            let fuxmap = trans_lock.read().unwrap();
            fuxmap.get(&fid).unwrap_or(&GATEWAY_ADDR.to_string()).clone() // todo: change if needed
        }
        None => GATEWAY_ADDR.to_string() // do i need to create the trans map ? 
    }
}

fn put(map: &TransactionMap, tid: Tid, fid: Fid, addr: Addr) {
    {
        let transmap = map.read().unwrap();
        
        if let Some(fux_lock) =transmap.get(&tid) {
            fux_lock.write().unwrap().insert(fid, addr);
            return;
        }
    }

    let mut transmap = map.write().unwrap();
    let fux_map = transmap.entry(tid).or_insert_with(|| Arc::new(RwLock::new(HashMap::new())));
    fux_map.write().unwrap().insert(fid, addr);
}

fn delete(map: &TransactionMap, tid: Tid) {
    let mut transmap = map.write().unwrap();
    transmap.remove_entry(&tid).unwrap(); // todo: verify
}
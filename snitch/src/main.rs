use dashmap::DashMap;
use dashmap::Entry::{Occupied, Vacant};
use std::env;
use std::sync::Arc;
use tokio::io::{AsyncBufReadExt, AsyncWriteExt, BufReader};
use tokio::net::{TcpListener, TcpStream};
use tokio::sync::watch;

type Addr = Arc<str>;
type Fid = String;
type Tid = String;

type FunctionMap = DashMap<Fid, watch::Sender<Option<Addr>>>;
type TransactionMap = DashMap<Tid, Arc<FunctionMap>>;

struct Snitch {
    tmap: TransactionMap,
    default_addr: Addr,
}

impl Snitch {
    fn new(default_address: impl Into<Addr>) -> Self {
        Snitch {
            tmap: DashMap::new(),
            default_addr: default_address.into(),
        }
    }
}

async fn resolve(snitch: &Snitch, tid: &str, fid: &str) -> Addr {
    let fuxmap = match snitch.tmap.entry(tid.to_owned()) {
        Occupied(entry) => entry.get().clone(),
        Vacant(entry) => {
            entry.insert(Arc::new(FunctionMap::new()));
            return snitch.default_addr.clone();
        }
    };

    let mut rx = match fuxmap.entry(fid.to_owned()) {
        Occupied(entry) => entry.get().subscribe(),
        Vacant(entry) => {
            let (sx, rx) = watch::channel(None);
            entry.insert(sx);
            rx
        }
    };

    loop {
        if let Some(addr) = rx.borrow().clone() {
            return addr;
        }

        if rx.changed().await.is_err() {
            return snitch.default_addr.clone();
        }
    }
}


fn put(snitch: &Snitch, tid: &str, fid: &str, addr: Addr) {
    let fuxmap = snitch.tmap.entry(tid.to_owned())
        .or_insert_with(|| Arc::new(FunctionMap::new()))
        .clone();

    match fuxmap.entry(fid.to_owned()) {
        Occupied(entry) => {
            entry.get().send_replace(Some(addr));
        }
        Vacant(entry) => {
            let (sx, _) = watch::channel(Some(addr));
            entry.insert(sx);
        }
    };
}



#[tokio::main]
async fn main() -> std::io::Result<()> {
    let bind_addr = "0.0.0.0:46655".to_string();
    let default_address =
        env::var("GATEWAY_ADDR").unwrap_or_else(|_| "localhost:8080".to_string());

    let snitch = Arc::new(Snitch::new(default_address));

    let listener = TcpListener::bind(&bind_addr).await?;
    eprintln!("snitch listening: {bind_addr}");

    loop {
        let (socket, peer) = listener.accept().await?;
        let snitch = Arc::clone(&snitch);
        tokio::spawn(async move {
            if let Err(e) = handle_connection(socket, &snitch).await {
                eprintln!("connection {peer} ended with error: {e}");
            }
        });
    }
}


async fn handle_connection(socket: TcpStream, snitch: &Arc<Snitch>) -> std::io::Result<()> {
    let (reader, mut writer) = socket.into_split();
    let mut lines = BufReader::new(reader).lines();

    while let Some(line) = lines.next_line().await? {
        let line = line.trim();

        if line.is_empty() {
            continue;
        }

        let response = process_line(&snitch, line).await;
        writer.write_all(response.as_bytes()).await?;
        //writer.write_all(b"\n").await?;
    }

    println!("Closing connection.");

    Ok(())
}


async fn process_line(snitch: &Snitch, line: &str) -> String {
    let parts: Vec<&str> = line.trim().split_whitespace().collect();

    match parts.as_slice() {
        ["GET", tid, fid] => {
            let (tid, fid) = match (tid.parse::<Tid>(), fid.parse::<Fid>()) {
                (Ok(t), Ok(f)) => (t, f),
                _ => return "Guck you".to_string(),
            };
            let addr = resolve(snitch, &tid, &fid).await;
            format!("{addr}")
        }
        ["PUT", tid, fid, addr] => {
            let (tid, fid, addr) = match (tid.parse::<Tid>(), fid.parse::<Fid>(), addr.parse::<String>()) {
                (Ok(t), Ok(f), Ok(a)) => (t, f, a),
                _ => return format!("Fuck yOu"),
            };
            let addr = Arc::from(addr);
            put(snitch, &tid, &fid, addr);
            "OK".to_string()
        }
        _ => format!("fuck You"),
    }
}
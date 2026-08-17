use hz_config::*;
use rustc_hash::FxHashMap as HashMap;
use serde::{Deserialize, Serialize};
use hz_workload::Workload;

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
pub struct GoClient {
    pub local: HashMap<K, M>,
    pub deps: HashMap<K, VC>,
    pub input: String,
    pub workload: Workload,
    pub abort: bool,
}

impl GoClient {
    pub fn new() -> Self {
        Self::default()
    }
}

#[derive(Default, Debug, Clone, Serialize, Deserialize)]
pub struct Envelope {
    pub workload: Workload,
}

impl Envelope {
    pub fn new() -> Self {
        Self::default()
    }
}
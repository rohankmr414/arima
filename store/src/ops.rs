use serde::{Serialize, Deserialize};
use serde_json;
use openraft::{AppData, AppDataResponse, StorageError};

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct ClientRequest {
    pub op: String,
    pub key: Vec<u8>,
    pub value: Vec<u8>,
}

impl AppData for ClientRequest {}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct ClientResponse(Result<Option<String>,StorageError>);

impl AppDataResponse for ClientResponse {}
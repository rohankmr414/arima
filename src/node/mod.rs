use slog::Drain;
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};

use protobuf::Message as ProtobufMessage;
use crate::storage::*;
use raft::{prelude::*, StateRole};

use slog::{error, info, o};

// structure for raft node
struct ArimaNode {
    raft_group: Option<RawNode<ArimaStorage>>,

}
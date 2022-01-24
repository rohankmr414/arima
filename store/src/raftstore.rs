use std::fmt::Debug;
use std::ops::RangeBounds;
use openraft::storage::{RaftStorage, HardState, Snapshot, SnapshotMeta, InitialState};
use db::kv::*;
use openraft::{EffectiveMembership, LogId, StateMachineChanges, StorageError};
use openraft::raft::Entry;
use serde::{Serialize, Deserialize};
use serde_json;

use crate::ops::{ClientRequest, ClientResponse};

pub struct ArimaStore {
    pub db: KV
}

impl RaftStorage<ClientRequest, ClientResponse> for ArimaStore {
    type SnapshotData = Vec<u8>;
    async fn get_membership_config(&self) -> Result<EffectiveMembership, StorageError> {
        todo!()
    }
    async fn get_initial_state(&self) -> Result<InitialState, StorageError> {
        todo!()
    }
    async fn save_hard_state(&self, hs: &HardState) -> Result<(), StorageError> {
        todo!()
    }
    async fn read_hard_state(&self) -> Result<Option<HardState>, StorageError> {
        todo!()
    }
    async fn get_log_entries<RNG: RangeBounds<u64> + Clone + Debug + Send + Sync>(&self, range: RNG) -> Result<Vec<Entry<D>>, StorageError> {
        todo!()
    }
    async fn try_get_log_entries<RNG: RangeBounds<u64> + Clone + Debug + Send + Sync>(&self, range: RNG) -> Result<Vec<Entry<D>>, StorageError> {
        todo!()
    }
    async fn try_get_log_entry(&self, log_index: u64) -> Result<Option<Entry<D>>, StorageError> {
        todo!()
    }
    async fn first_id_in_log(&self) -> Result<Option<LogId>, StorageError> {
        todo!()
    }
    async fn first_known_log_id(&self) -> Result<LogId, StorageError> {
        todo!()
    }
    async fn last_id_in_log(&self) -> Result<LogId, StorageError> {
        todo!()
    }
    async fn last_applied_state(&self) -> Result<(LogId, Option<EffectiveMembership>), StorageError> {
        todo!()
    }
    async fn delete_logs_from<RNG: RangeBounds<u64> + Clone + Debug + Send + Sync>(&self, range: RNG) -> Result<(), StorageError> {
        todo!()
    }
    async fn append_to_log(&self, entries: &[&Entry<D>]) -> Result<(), StorageError> {
        todo!()
    }
    async fn apply_to_state_machine(&self, entries: &[&Entry<D>]) -> Result<Vec<R>, StorageError> {
        todo!()
    }
    async fn do_log_compaction(&self) -> Result<Snapshot<Self::SnapshotData>, StorageError> {
        todo!()
    }
    async fn begin_receiving_snapshot(&self) -> Result<Box<Self::SnapshotData>, StorageError> {
        todo!()
    }
    async fn finalize_snapshot_installation(&self, meta: &SnapshotMeta, snapshot: Box<Self::SnapshotData>) -> Result<StateMachineChanges, StorageError> {
        todo!()
    }
    async fn get_current_snapshot(&self) -> Result<Option<Snapshot<Self::SnapshotData>>, StorageError> {
        todo!()
    }
}

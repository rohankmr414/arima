use raft::{Config}

impl Config {
    pub fn new(id: u64, peers: Vec<u64>, election_tick: usize, heartbeat_tick: usize) -> Config {
        Config {
            id: id,
            peers: peers,
            election_tick: election_tick,
            heartbeat_tick: heartbeat_tick,
            ..Default::default()
        }
    }
    pub fn min_election_tick(&self) -> usize {
        self.election_tick
    }
    pub fn max_election_tick(&self) -> usize {
        self.election_tick * 2
    }

    // function to run validations against the config.
    pub fn validate(&self) -> Result<(), String> {
        if self.id == 0 {
            return Err("id cannot be 0".to_string());
        }
        if self.peers.is_empty() {
            return Err("peers cannot be empty".to_string());
        }
        if self.peers.contains(&self.id) {
            return Err(format!("id cannot be in peers: {}", self.id));
        }
        if self.peers.len() != self.peers.iter().unique().count() {
            return Err("peers cannot contain duplicates".to_string());
        }
        if self.election_tick == 0 {
            return Err("election_tick cannot be 0".to_string());
        }
        if self.heartbeat_tick == 0 {
            return Err("heartbeat_tick cannot be 0".to_string());
        }
        if self.heartbeat_tick >= self.election_tick {
            return Err("heartbeat_tick must be less than election_tick".to_string());
        }
        Ok(())
    }
}
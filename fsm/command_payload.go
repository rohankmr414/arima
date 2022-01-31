package fsm

// CommandPayload is payload sent by system when calling raft.Apply(cmd []byte, timeout time.Duration)
type CommandPayload struct {
	Operation string
	Key       []byte
	Value     []byte
}
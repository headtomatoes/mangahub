package tcp

type Message struct {
	Type string                 `json:"type"` // basic routing based on type field
	Data map[string]interface{} `json:"data"` // flexible data payload
}

package response

type ResponseType string

const (
	Success ResponseType = "success"
	Error   ResponseType = "error"
)

type Response struct {
	Type    ResponseType `json:"type"`
	Message string       `json:"message"`
	Data    interface{}  `json:"data,omitempty"`
}

package structs

type MessageRequest struct {
	Message string `json:"message"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type WebSocketMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

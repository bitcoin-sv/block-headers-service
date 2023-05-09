package http

// WebhookRequest defines a request body for webhook registration.
type WebhookRequest struct {
	Url          string 		`json:"url"`
	RequiredAuth RequiredAuth	`json:"requiredAuth"`
}

// RequiredAuth defines an auth information for webhook registration.
type RequiredAuth struct {
	Type	string `json:"type"`
	Token	string `json:"token"`
	Header	string `json:"header"`
}

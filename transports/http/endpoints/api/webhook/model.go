package webhook

// Request defines a request body for webhook registration.
type Request struct {
	URL          string       `json:"url"`
	RequiredAuth RequiredAuth `json:"requiredAuth"`
}

// RequiredAuth defines an auth information for webhook registration.
type RequiredAuth struct {
	Type   string `json:"type"`
	Token  string `json:"token"`
	Header string `json:"header"`
}

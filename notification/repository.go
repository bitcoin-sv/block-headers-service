package notification

// Webhooks is an interface which represents methods performed on registered_webhooks table in defined storage.
type Webhooks interface {
	AddWebhookToDatabase(token *Webhook) error
	DeleteWebhookByURL(url string) error
	GetWebhookByURL(url string) (*Webhook, error)
	GetAllWebhooks() ([]*Webhook, error)
	UpdateWebhook(w *Webhook) error
}

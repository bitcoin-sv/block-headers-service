package notification

// Event represents event to notify with.
type Event any

// Channel is a component representing channel of communication ex. http request to webhook, websocket etc.
type Channel interface {
	// Notify send event notification.
	Notify(Event)
}

// Notifier is representing component that can be used to notify clients about important events.
type Notifier struct {
	channels []Channel
}

// NewNotifier create Notifier.
func NewNotifier() *Notifier {
	return &Notifier{
		channels: make([]Channel, 0),
	}
}

// AddChannel register communication channel in notifier.
func (n *Notifier) AddChannel(ch Channel) {
	n.channels = append(n.channels, ch)
}

// Notify send event notification via registered channels.
func (n *Notifier) Notify(event any) {
	for _, ch := range n.channels {
		go ch.Notify(event)
	}
}

package events
// exchanges
const (
	ExchangeNotifications = "notifications.events" 
)
// routing keys
const (
	RoutingUserRegistered = "notifications.user.registered"
)
// queues
const (
	QueueUserRegisteredEmailNotifier = "notifications.user.registered.email_notifier"
	QueueUserRegisteredPushNotifier = "notifications.user.registered.push_notifier"
)
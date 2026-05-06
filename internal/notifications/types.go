package notifications
import (
	"github.com/resend/resend-go/v2"
)

const (
	ChannelEmail = "email"
	ChannelPush  = "push"
)

type EmailService struct {
	client   *resend.Client
	fromEmail string
}
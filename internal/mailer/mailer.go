package mailer

import "embed"

const (
	UserWelcomeTemplate      = "templates/activate_user_temp.tmpl"
	AdminUserWelcomeTemplate = "templates/activate_admin_user_temp.tmpl"
	maxRetries               = 3
	fromUser                 = "Eye Of"
)

//go:embed templates
var FS embed.FS

type Client interface {
	Send(templateFile, username, email string, data any) error
}

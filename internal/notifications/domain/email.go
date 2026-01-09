package domain

type Email struct {
	To      string
	Subject string
	Text    string
	HTML    string
}

type EmailService interface {
	Send(email *Email) error
}

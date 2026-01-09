package jobs

import "newsletter/internal/notifications/domain"

type SendEmailJob struct {
	Email   domain.Email
	Service domain.EmailService
}

func (job *SendEmailJob) Process() error {
	err := job.Service.Send(&job.Email)
	return err
}

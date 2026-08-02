package sender

type Sender interface {
	Send(payload any) (Status, error)
}

package ddns

type UpdateStatus int

const (
	UpdaterStatusOk UpdateStatus = iota
	UpdaterStatusUnchanged
)

func (s UpdateStatus) String() string {
	switch s {
	case UpdaterStatusOk:
		return "Ok"
	case UpdaterStatusUnchanged:
		return "Unchanged"
	default:
		return "Unknown"
	}
}

type UpdateResult struct {
	Status UpdateStatus
	IPv4   string
	IPv6   string
}

type Updater interface {
	Update() (UpdateResult, error)
}

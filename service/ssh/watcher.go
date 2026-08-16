package ssh

import (
	"context"
	"fmt"
	"keeper/pkg/eventbus"
	"log/slog"
	"regexp"
	"time"

	"github.com/coreos/go-systemd/v22/sdjournal"
)

const (
	journalStepRetryDelay = 1 * time.Second
	journalWaitTimeout    = 5 * time.Second
)

var (
	reSuccess = regexp.MustCompile(`Accepted \S+ for (\S+) from (\S+) port`)
	reFailure = regexp.MustCompile(`Failed \S+ for (?:invalid user )?(\S+) from (\S+) port`)
	reInvalid = regexp.MustCompile(`Invalid user (\S+) from (\S+) port`)

	reDisconnect   = regexp.MustCompile(`Disconnected from user (\S+) (\S+) port`)
	rePreauthClose = regexp.MustCompile(`Connection closed by (\S+) port \d+ \[preauth\]`)

	reMaxTries    = regexp.MustCompile(`Maximum authentication attempts exceeded for (?:invalid user )?(\S+) from (\S+) port`)
	reInvalidKey  = regexp.MustCompile(`userauth_pubkey: key type \S+ not in PubkeyAcceptedAlgorithms`)
	reSudoSession = regexp.MustCompile(`pam_unix\(sudo:session\): session opened for user (\S+) by (\S+)`)
)

type EventType int

const (
	LoginOk EventType = iota
	LoginFailed
	MaxTriesExceeded
	InvalidKeyAlgorithm
	Disconnect
	PreAuthDisconnect
	SudoEscalation
)

type Event struct {
	Type   EventType
	User   string
	IP     string
	Detail string
}

type Watcher struct {
	eventBus *eventbus.EventBus[Event]
}

func NewWatcher() *Watcher {
	return &Watcher{
		eventBus: eventbus.New[Event](),
	}
}

func (w *Watcher) EventBroadcaster() eventbus.Broadcaster[Event] {
	return w.eventBus
}

func (w *Watcher) Run(ctx context.Context) error {
	j, err := sdjournal.NewJournal()
	if err != nil {
		return fmt.Errorf("failed to open journal: %w", err)
	}
	defer j.Close()

	if err := j.AddMatch("SYSLOG_IDENTIFIER=sshd"); err != nil {
		return fmt.Errorf("match error: %w", err)
	}
	if err := j.AddDisjunction(); err != nil { // OR operator
		return fmt.Errorf("match disjunction error: %w", err)
	}
	if err := j.AddMatch("SYSLOG_IDENTIFIER=sudo"); err != nil {
		return fmt.Errorf("match error: %w", err)
	}

	if err := j.SeekTail(); err != nil {
		return fmt.Errorf("failed to seek tail: %w", err)
	}
	if _, err := j.Previous(); err != nil {
		return fmt.Errorf("failed to step back: %w", err)
	}

	log := slog.With("component", "SSHWatcher")
	log.Info("Listening for live SSH & Sudo systemd events...")

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		n, err := j.Next()
		if err != nil {
			log.Error("Error stepping through journal", "error", err)
			time.Sleep(journalStepRetryDelay)
			continue
		}

		// end reached
		if n == 0 {
			j.Wait(journalWaitTimeout)
			continue
		}

		msg, err := j.GetData("MESSAGE")
		if err != nil {
			log.Debug("Failed to get MESSAGE field from journal entry", "error", err)
			continue
		}

		if event, parsed := parseMessage(msg); parsed {
			w.eventBus.Publish(event)
		}
	}
}

func parseMessage(msg string) (Event, bool) {
	if m := reSuccess.FindStringSubmatch(msg); len(m) == 3 {
		return Event{Type: LoginOk, User: m[1], IP: m[2]}, true
	}

	if m := reMaxTries.FindStringSubmatch(msg); len(m) == 3 {
		return Event{Type: MaxTriesExceeded, User: m[1], IP: m[2]}, true
	}

	if m := reFailure.FindStringSubmatch(msg); len(m) == 3 {
		return Event{Type: LoginFailed, User: m[1], IP: m[2]}, true
	}
	if m := reInvalid.FindStringSubmatch(msg); len(m) == 3 {
		return Event{Type: LoginFailed, User: m[1], IP: m[2]}, true
	}

	if reInvalidKey.MatchString(msg) {
		return Event{Type: InvalidKeyAlgorithm, User: "unknown", IP: "unknown", Detail: msg}, true
	}

	if m := reDisconnect.FindStringSubmatch(msg); len(m) == 3 {
		return Event{Type: Disconnect, User: m[1], IP: m[2]}, true
	}

	if m := rePreauthClose.FindStringSubmatch(msg); len(m) == 2 {
		return Event{Type: PreAuthDisconnect, User: "none", IP: m[1]}, true
	}

	if m := reSudoSession.FindStringSubmatch(msg); len(m) == 3 {
		return Event{
			Type:   SudoEscalation,
			User:   m[2],
			IP:     "local",
			Detail: fmt.Sprintf("Escalated to %s", m[1]),
		}, true
	}

	return Event{}, false
}

package ukaz

import (
	"strings"

	log "log/slog"

	"ukaz/pkg/proto"
)

// requestDeadlinesFromGovernor sends GET:DEADLINES to the governor and records the
// requester so we can print and reply when OK:DEADLINES arrives.
func (u *Ukaz) requestDeadlinesFromGovernor(req *proto.Request) {
	msg := req.Msg
	u.deadlinesMu.Lock()
	u.pendingDeadlinesFrom = msg.From
	u.deadlinesMu.Unlock()

	args := msg.Args
	if len(args) > 0 {
		_ = u.client.Send(governorNodeID, "GET", "DEADLINES", args[0])
	} else {
		_ = u.client.Send(governorNodeID, "GET", "DEADLINES")
	}
	log.Debug("GET DEADLINES", "from", msg.From, "requested governor")
}

// handleDeadlinesResponse is called when we receive OK:DEADLINES from the governor.
// It prints each event line and replies to the pending requester.
func (u *Ukaz) handleDeadlinesResponse(events []string) {
	u.deadlinesMu.Lock()
	replyTo := u.pendingDeadlinesFrom
	u.pendingDeadlinesFrom = ""
	u.deadlinesMu.Unlock()

	if u.printer != nil {
		// Header in 9x9 (mode 0 on ND77)
		_ = u.printer.SetPrintMode(0)
		_ = u.printer.PrintLineWithSetup("--- Deadlines ---")
		_ = u.printer.FeedLines(1)
		// Deadlines in 7x9 (mode 1 on ND77)
		_ = u.printer.SetPrintMode(1)
		for _, ev := range events {
			line := strings.ReplaceAll(ev, "|", "  ")
			if err := u.printer.PrintLineWithSetup(line); err != nil {
				log.Warn("print deadline line", "err", err, "line", line)
			}
		}
		_ = u.printer.FeedLines(1)
		// Footer in 9x9
		_ = u.printer.SetPrintMode(0)
		_ = u.printer.PrintLineWithSetup("---")
		_ = u.printer.FeedLines(4) // blank space then full cut
		_ = u.printer.FullCut()
		_ = u.printer.SetPrintMode(1) // restore 7x9 default
	}

	if replyTo != "" {
		_ = u.client.Send(replyTo, "OK", "DEADLINES", events...)
	}
	log.Info("DEADLINES printed", "count", len(events), "replyTo", replyTo)
}

package ukaz

import (
	"time"

	log "log/slog"

	"ukaz/pkg/proto"
)

// requestSystemStatus sends GET:UPTIME to LUCH, VERTEX, ACHTUNG, GOVERNOR,
// collects OK:UPTIME responses for statusCollectTimeout, then prints the report.
func (u *Ukaz) requestSystemStatus(req *proto.Request) {
	msg := req.Msg
	u.statusMu.Lock()
	if u.statusCollecting {
		u.statusMu.Unlock()
		log.Warn("STATUS already in progress", "from", msg.From)
		req.Reply("ERR", "STATUS", "BUSY")
		return
	}
	u.statusCollecting = true
	u.statusResults = make(map[string]string)
	u.statusRequester = msg.From
	u.statusMu.Unlock()

	for _, node := range statusNodes {
		_ = u.client.Send(node, "GET", "UPTIME")
	}
	log.Debug("GET STATUS", "from", msg.From, "queried", statusNodes)

	go func() {
		time.Sleep(statusCollectTimeout)
		u.finishSystemStatus()
	}()
}

// handleUptimeResponse records OK:UPTIME from a node when collecting system status.
func (u *Ukaz) handleUptimeResponse(from string, args []string) {
	u.statusMu.Lock()
	defer u.statusMu.Unlock()
	if !u.statusCollecting {
		return
	}
	uptime := "--"
	if len(args) > 0 {
		uptime = args[0]
	}
	u.statusResults[from] = uptime
}

// finishSystemStatus prints collected uptimes and clears status collection.
func (u *Ukaz) finishSystemStatus() {
	u.statusMu.Lock()
	if !u.statusCollecting {
		u.statusMu.Unlock()
		return
	}
	results := make(map[string]string)
	for k, v := range u.statusResults {
		results[k] = v
	}
	requester := u.statusRequester
	u.statusCollecting = false
	u.statusResults = nil
	u.statusRequester = ""
	u.statusMu.Unlock()

	if u.printer != nil {
		// Header in 9x9
		_ = u.printer.SetPrintMode(0)
		_ = u.printer.PrintLineWithSetup("--- System Status ---")
		_ = u.printer.FeedLines(1)
		// Lines in 7x9
		_ = u.printer.SetPrintMode(1)
		for _, node := range statusNodes {
			uptime := results[node]
			if uptime == "" {
				uptime = "--"
			}
			_ = u.printer.PrintLineWithSetup(node + ": " + uptime)
		}
		_ = u.printer.FeedLines(1)
		// Footer in 9x9
		_ = u.printer.SetPrintMode(0)
		_ = u.printer.PrintLineWithSetup("---")
		_ = u.printer.FeedLines(8)
		_ = u.printer.FullCut()
		_ = u.printer.SetPrintMode(1)
	}

	if requester != "" {
		parts := make([]string, 0, len(statusNodes)*2)
		for _, node := range statusNodes {
			uptime := results[node]
			if uptime == "" {
				uptime = "--"
			}
			parts = append(parts, node, uptime)
		}
		_ = u.client.Send(requester, "OK", "STATUS", parts...)
	}
	log.Info("STATUS printed", "replyTo", requester)
}

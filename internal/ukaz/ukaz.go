package ukaz

import (
	"strings"
	"sync"

	log "log/slog"
	"time"

	"ukaz/pkg/proto"
)

const governorNodeID = "GOVERNOR"

// Known nodes queried for system status (GET:UPTIME).
var statusNodes = []string{"LUCH", "VERTEX", "ACHTUNG", "GOVERNOR"}

const statusCollectTimeout = 2 * time.Second

// Printer is the minimal interface needed to print when requested.
// If nil, PRINT commands reply ERR NO_DEVICE.
type Printer interface {
	Init() error
	PrintLineWithSetup(text string) error
	SetPrintMode(mode byte) error // 7x9 = 0, 9x9 = 1 (Font B)
	FeedLines(n int) error        // feed n lines (e.g. for cutting); no-op if not supported
	PartialCut() error             // partial (perforation) cut
	FullCut() error                // full cut after feeding
	Close() error
}

// Ukaz is the ND77 matrix printer driver node.
type Ukaz struct {
	client   *proto.Client
	printer  Printer
	bootedAt time.Time

	deadlinesMu          sync.Mutex
	pendingDeadlinesFrom string // who asked for GET:DEADLINES; we reply when governor answers

	statusMu         sync.Mutex
	statusCollecting bool
	statusResults   map[string]string // node -> uptime
	statusRequester string
}

// New creates the driver. Pass a non-nil printer when a serial device is configured.
func New(client *proto.Client, printer Printer) *Ukaz {
	return &Ukaz{
		client:   client,
		printer:  printer,
		bootedAt: time.Now(),
	}
}

// Cmd dispatches an incoming request by verb.
//
//	PING PING       -> PONG PONG
//	GET  UPTIME     -> OK UPTIME <dur>
//	PRINT LINE <text> -> OK PRINT | ERR PRINT <reason>
//	PRINT INIT       -> OK PRINT | ERR PRINT <reason>
func (u *Ukaz) Cmd(req *proto.Request) {
	msg := req.Msg
	log.Debug("CMD", "from", msg.From, "verb", msg.Verb, "noun", msg.Noun, "args", msg.Args)

	switch msg.Verb {
	case "ERR", "PONG":
		log.Debug("IGNORE", "verb", msg.Verb, "noun", msg.Noun, "from", msg.From)
		return
	case "OK":
		if msg.Noun == "DEADLINES" && msg.From == governorNodeID {
			u.handleDeadlinesResponse(msg.Args)
			return
		}
		if msg.Noun == "UPTIME" {
			u.handleUptimeResponse(msg.From, msg.Args)
			return
		}
		log.Debug("IGNORE", "verb", msg.Verb, "noun", msg.Noun, "from", msg.From)
		return
	case "PING":
		req.Reply("PONG", "PONG")
	case "GET":
		u.cmdGet(req)
	case "PRINT":
		u.cmdPrint(req)
	default:
		log.Warn("UNKNOWN VERB", "verb", msg.Verb, "from", msg.From)
		req.Reply("ERR", "VERB")
	}
}

func (u *Ukaz) cmdGet(req *proto.Request) {
	msg := req.Msg
	switch msg.Noun {
	case "UPTIME":
		uptime := time.Since(u.bootedAt).Truncate(time.Second)
		log.Debug("GET UPTIME", "uptime", uptime, "from", msg.From)
		req.Reply("OK", "UPTIME", uptime.String())
	case "DEADLINES":
		u.requestDeadlinesFromGovernor(req)
	case "STATUS":
		u.requestSystemStatus(req)
	default:
		log.Warn("UNKNOWN NOUN", "noun", msg.Noun, "from", msg.From)
		req.Reply("ERR", "NOUN")
	}
}

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

func (u *Ukaz) cmdPrint(req *proto.Request) {
	msg := req.Msg
	if u.printer == nil {
		log.Warn("PRINT no device", "from", msg.From)
		req.Reply("ERR", "PRINT", "NO_DEVICE")
		return
	}
	switch msg.Noun {
	case "INIT":
		if err := u.printer.Init(); err != nil {
			log.Error("PRINT INIT failed", "err", err, "from", msg.From)
			req.Reply("ERR", "PRINT", err.Error())
			return
		}
		log.Info("PRINT INIT", "from", msg.From)
		req.Reply("OK", "PRINT")
	case "LINE":
		line := strings.Join(msg.Args, ":")
		if err := u.printer.PrintLineWithSetup(line); err != nil {
			log.Error("PRINT LINE failed", "err", err, "from", msg.From)
			req.Reply("ERR", "PRINT", err.Error())
			return
		}
		log.Info("PRINT LINE", "line", line, "from", msg.From)
		req.Reply("OK", "PRINT")
	default:
		log.Warn("UNKNOWN NOUN", "noun", msg.Noun, "from", msg.From)
		req.Reply("ERR", "NOUN")
	}
}

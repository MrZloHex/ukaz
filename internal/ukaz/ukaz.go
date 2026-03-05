package ukaz

import (
	"sync"
	"time"

	log "log/slog"

	"ukaz/pkg/printer"
	"ukaz/pkg/proto"
)

// Ukaz is the ND77 matrix printer driver node.
type Ukaz struct {
	client   *proto.Client
	printer  printer.Device
	bootedAt time.Time

	deadlinesMu          sync.Mutex
	pendingDeadlinesFrom string // who asked for GET:DEADLINES; we reply when governor answers

	statusMu         sync.Mutex
	statusCollecting bool
	statusResults   map[string]string // node -> uptime
	statusRequester string
}

// New creates the driver. Pass a non-nil printer when a serial device is configured.
func New(client *proto.Client, dev printer.Device) *Ukaz {
	return &Ukaz{
		client:   client,
		printer:  dev,
		bootedAt: time.Now(),
	}
}

// Cmd dispatches an incoming request by verb.
//
//	PING PING       -> PONG PONG
//	GET  UPTIME     -> OK UPTIME <dur>
//	GET  DEADLINES  -> queries governor, prints, replies
//	GET  STATUS     -> queries all nodes, prints uptimes, replies
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

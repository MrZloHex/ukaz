package ukaz

import (
	"time"

	log "log/slog"

	"ukaz/pkg/proto"
)

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

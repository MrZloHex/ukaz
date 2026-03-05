package ukaz

import (
	"strings"

	log "log/slog"

	"ukaz/pkg/proto"
)

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

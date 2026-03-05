package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
	log "log/slog"

	cli "github.com/spf13/pflag"

	"ukaz/internal/ukaz"
	"ukaz/pkg/printer"
	"ukaz/pkg/proto"
)

var logLevelMap = map[string]log.Level{
	"debug": log.LevelDebug,
	"info":  log.LevelInfo,
	"warn":  log.LevelWarn,
	"error": log.LevelError,
}

func main() {
	url := cli.StringP("url", "u", "ws://localhost:8092", "Url of hub")
	logLevel := cli.StringP("log", "l", "info", "Log level")
	serialPort := cli.StringP("serial", "s", "", "Serial port for ND77 (e.g. /dev/ttyUSB0). Omit to run without printer.")
	baud := cli.IntP("baud", "b", 9600, "Serial baud rate for ND77")
	widthMm := cli.Float64P("width", "w", 69.5, "Paper width in mm: 69.5 (40/30 chars) or 57.5 (33/25 chars)")
	cli.Parse()

	log.SetDefault(log.New(tint.NewHandler(os.Stdout, &tint.Options{
		Level: logLevelMap[*logLevel],
	})))

	client := proto.New("UKAZ", *url,
		proto.WithReconnect(5*time.Second),
	)

	var prn ukaz.Printer
	if *serialPort != "" {
		p, err := printer.Open(*serialPort, *baud)
		if err != nil {
			log.Error("Failed to open serial port", "port", *serialPort, "err", err)
			os.Exit(1)
		}
		p.SetPaperWidth(*widthMm)
		prn = p
		log.Info("ND77 printer attached", "port", *serialPort, "baud", *baud, "widthMm", *widthMm)
		defer func() {
			if err := prn.Close(); err != nil {
				log.Warn("Printer close", "err", err)
			}
		}()
	} else {
		log.Info("No serial port; PRINT commands will reply ERR NO_DEVICE")
	}

	drv := ukaz.New(client, prn)

	client.Handle("*", func(req *proto.Request) {
		if req.Msg.To != client.NodeID() {
			return
		}
		drv.Cmd(req)
	})

	log.Info("BOOTING UP", "url", *url)

	if err := client.Connect(context.Background()); err != nil {
		log.Error("Failed to connect", "err", err)
		os.Exit(1)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Info("SHUTTING DOWN")
	client.Close()
}

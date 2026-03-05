███╗   ███╗ ██████╗ ███╗   ██╗ ██████╗ ██╗     ██╗████████╗██╗  ██╗
████╗ ████║██╔═══██╗████╗  ██║██╔═══██╗██║     ██║╚══██╔══╝██║  ██║
██╔████╔██║██║   ██║██╔██╗ ██║██║   ██║██║     ██║   ██║   ███████║
██║╚██╔╝██║██║   ██║██║╚██╗██║██║   ██║██║     ██║   ██║   ██╔══██║
██║ ╚═╝ ██║╚██████╔╝██║ ╚████║╚██████╔╝███████╗██║   ██║   ██║  ██║
╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝ ╚══════╝╚═╝   ╚═╝   ╚═╝  ╚═╝


  ░▒▓█ _ukaz_ █▓▒░  
  ND77 matrix printer driver - dot-by-dot, line-by-line.  

  ───────────────────────────────────────────────────────────────  
  ▓ OVERVIEW  
  **_ukaz_** is a matrix printer (ND77) driver microservice written in **Go**.  
  It connects to a _concentrator_ hub via WebSocket,  
  accepting commands to drive the printer.  
  When a serial port is configured, it accepts **PRINT** commands  
  and sends ESC/POS to the ND77 (receipt station, line-by-line).  

  ───────────────────────────────────────────────────────────────  
  ▓ ARCHITECTURE  
  ▪ **RUNTIME**: Go 1.25  
  ▪ **TRANSPORT**: WebSocket (gorilla/websocket) via pkg/proto client  
  ▪ **PRINTER**: Serial (go.bug.st/serial), ESC/POS to ND77  
  ▪ **NODE ID**: UKAZ  

  ───────────────────────────────────────────────────────────────  
  ▓ FEATURES  
  ▪ Ping/pong health check  
  ▪ Uptime reporting  
  ▪ Print on request (PRINT:LINE, PRINT:INIT) when -s is set  
  ▪ Auto-reconnect on WebSocket disconnect  
  ▪ Graceful shutdown on SIGINT/SIGTERM  

  ───────────────────────────────────────────────────────────────  
  ▓ BUILD & RUN  
  ```sh  
  go build -o bin/ukaz ./cmd/ukaz  
  ./bin/ukaz -u ws://localhost:8092 -l info  
  ./bin/ukaz -u ws://localhost:8092 -s /dev/ttyUSB0 -b 9600  
  ```

  Flags:  
  ▪ `-u`  WebSocket hub URL  (default: ws://localhost:8092)  
  ▪ `-l`  Log level: debug, info, warn, error  (default: info)  
  ▪ `-s`  Serial port for ND77 (e.g. /dev/ttyUSB0). Omit to run without printer.  
  ▪ `-b`  Serial baud rate for ND77  (default: 9600)  

  ───────────────────────────────────────────────────────────────  
  ▓ PROTOCOL  
  Packet format:  <TO>:<VERB>:<NOUN>[:<ARGS>...]:<FROM>  

  Responses:  OK:<NOUN>[:ARGS]  or  ERR:<REASON>[:ARGS]  

  ─── PING ───  
  PING:PING                        -> PONG:PONG  

  ─── GET ───  
  GET:UPTIME                       -> OK:UPTIME:<duration>  

  ─── PRINT ───  
  PRINT:INIT                       -> OK:PRINT  |  ERR:PRINT:<reason>  
  PRINT:LINE[:<text>]              -> OK:PRINT  |  ERR:PRINT:<reason>  
  (Args are joined with ":" to form the line. No -s: ERR:PRINT:NO_DEVICE.)  

  ───────────────────────────────────────────────────────────────  
  ▓ FINAL WORDS  
  Ready to print when you are.  

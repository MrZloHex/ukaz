// Package printer provides minimal ESC/POS output for the ND77 matrix printer
// over a serial port. Used by the ukaz monolith node to print on request.
package printer

import (
	"io"
	"strings"
	"time"

	"go.bug.st/serial"
)

const (
	esc = 0x1B
	lf  = 0x0A
)

// ND77 receipt station (ESC c 0 n).
const stationReceipt = 0x01

// Print mode (ESC ! n): font selection for line width.
// 7x9 = 40/33 chars per line, 9x9 = 30/25 chars per line (69.5/57.5 mm).
const (
	Mode7x9 = 0
	Mode9x9 = 1 << 0 // Font B (ND77_MODE_FONT_B)
)

// Paper width in mm. Determines max characters per line (7x9 and 9x9 fonts).
const (
	Width69_5mm = 69.5 // 40 chars (7x9), 30 chars (9x9)
	Width57_5mm = 57.5 // 33 chars (7x9), 25 chars (9x9)
)

// Printer sends ESC/POS commands to an ND77 over a serial port.
type Printer struct {
	port      serial.Port
	widthMm   float64 // paper width: 69.5 or 57.5
	chars7x9  int
	chars9x9  int
}

// Open opens the serial port and returns a Printer. Baud is typically 9600.
// Paper width defaults to 69.5 mm; call SetPaperWidth to use 57.5 mm.
func Open(portPath string, baud int) (*Printer, error) {
	mode := &serial.Mode{BaudRate: baud}
	port, err := serial.Open(portPath, mode)
	if err != nil {
		return nil, err
	}
	p := &Printer{port: port}
	p.SetPaperWidth(Width69_5mm)
	return p, nil
}

// SetPaperWidth sets the paper width in mm (69.5 or 57.5). Used for character limits.
func (p *Printer) SetPaperWidth(mm float64) {
	p.widthMm = mm
	switch {
	case mm <= 57.5:
		p.chars7x9 = 33
		p.chars9x9 = 25
	default:
		p.chars7x9 = 40
		p.chars9x9 = 30
	}
}

// PaperWidthMm returns the configured paper width in mm.
func (p *Printer) PaperWidthMm() float64 { return p.widthMm }

// Chars7x9 returns max characters per line for 7x9 font at current paper width.
func (p *Printer) Chars7x9() int { return p.chars7x9 }

// Chars9x9 returns max characters per line for 9x9 font at current paper width.
func (p *Printer) Chars9x9() int { return p.chars9x9 }

// Init sends ESC @ to reset the printer to defaults.
func (p *Printer) Init() error {
	_, err := p.port.Write([]byte{esc, '@'})
	return err
}

// SelectReceipt sends ESC c 0 1 to select the receipt station.
func (p *Printer) SelectReceipt() error {
	_, err := p.port.Write([]byte{esc, 'c', 0, stationReceipt})
	return err
}

// SetPrintMode sets the font for subsequent text (ESC ! n).
// Use Mode7x9 (default) or Mode9x9 for header/larger text.
func (p *Printer) SetPrintMode(mode byte) error {
	if p.port == nil {
		return nil
	}
	_, err := p.port.Write([]byte{esc, '!', mode})
	return err
}

// PrintLine sends text followed by LF. It does not call Init or SelectReceipt;
// the caller should ensure those were sent once (e.g. at startup or first print).
func (p *Printer) PrintLine(text string) error {
	text = strings.TrimSuffix(text, "\n")
	if _, err := p.port.Write([]byte(text)); err != nil {
		return err
	}
	_, err := p.port.Write([]byte{lf})
	return err
}

// PrintLineWithSetup ensures receipt station is selected then prints the line.
// Safe to call for every line; Init is not repeated.
func (p *Printer) PrintLineWithSetup(text string) error {
	if err := p.SelectReceipt(); err != nil {
		return err
	}
	return p.PrintLine(text)
}

// FeedLines sends ESC d n to feed n lines (e.g. to leave blank space for cutting).
func (p *Printer) FeedLines(n int) error {
	if p.port == nil || n <= 0 {
		return nil
	}
	if n > 255 {
		n = 255
	}
	_, err := p.port.Write([]byte{esc, 'd', byte(n)})
	return err
}

// PartialCut sends ESC m to perform a partial (perforation) cut.
func (p *Printer) PartialCut() error {
	if p.port == nil {
		return nil
	}
	_, err := p.port.Write([]byte{esc, 'm'})
	return err
}

// FullCut sends ESC i to perform a full cut.
func (p *Printer) FullCut() error {
	if p.port == nil {
		return nil
	}
	_, err := p.port.Write([]byte{esc, 'i'})
	return err
}

// Close closes the serial port.
func (p *Printer) Close() error {
	if p.port == nil {
		return nil
	}
	err := p.port.Close()
	p.port = nil
	return err
}

// Writer returns the underlying io.Writer for raw ESC/POS (e.g. from nd77 lib).
func (p *Printer) Writer() io.Writer { return p.port }

// Drain waits for the port's output buffer to be sent (best-effort).
func (p *Printer) Drain() error {
	if p.port == nil {
		return nil
	}
	if d, ok := p.port.(interface{ Drain() error }); ok {
		return d.Drain()
	}
	time.Sleep(50 * time.Millisecond)
	return nil
}

#!/usr/bin/env python3

import sys
import signal
import argparse

try:
    import serial
except ImportError:
    print("Missing pyserial. Install it with: pip install pyserial", file=sys.stderr)
    raise SystemExit(1)


RESET  = "\033[0m"
RED    = "\033[31m"
YELLOW = "\033[33m"
GREEN  = "\033[32m"
BLUE   = "\033[34m"
CYAN   = "\033[36m"
MAGENTA = "\033[35m"


def colorize_line(line: str) -> str:
    if "\x1b[" in line:
        return line

    stripped = line.lstrip()

    if stripped.startswith("E "):
        return RED + line + RESET
    if stripped.startswith("W "):
        return YELLOW + line + RESET
    if stripped.startswith("I "):
        return GREEN + line + RESET
    if stripped.startswith("D "):
        return CYAN + line + RESET
    if stripped.startswith("V "):
        return BLUE + line + RESET

    if "error" in stripped.lower():
        return RED + line + RESET
    if "warn" in stripped.lower():
        return YELLOW + line + RESET
    if "connected" in stripped.lower() or "got ip" in stripped.lower() or "ready" in stripped.lower():
        return GREEN + line + RESET

    return line


def main() -> int:
    parser = argparse.ArgumentParser(description="Simple colored serial monitor")
    parser.add_argument("port", help="Serial port, for example /dev/ttyUSB0")
    parser.add_argument("-b", "--baudrate", type=int, default=115200, help="Baudrate")
    parser.add_argument("--raw", action="store_true", help="Do not add own colors, only pass bytes through")
    args = parser.parse_args()

    try:
        ser = serial.Serial(args.port, args.baudrate, timeout=0.1)
    except Exception as e:
        print(f"Cannot open serial port {args.port}: {e}", file=sys.stderr)
        return 1

    running = True

    def handle_sigint(signum, frame):
        nonlocal running
        running = False

    signal.signal(signal.SIGINT, handle_sigint)

    print(f"{MAGENTA}Opened {args.port} @ {args.baudrate}{RESET}", file=sys.stderr)

    partial = bytearray()

    try:
        while running:
            chunk = ser.read(1024)
            if not chunk:
                continue

            if args.raw:
                sys.stdout.buffer.write(chunk)
                sys.stdout.buffer.flush()
                continue

            partial.extend(chunk)

            while True:
                nl_pos = partial.find(b"\n")
                if nl_pos == -1:
                    break

                line_bytes = partial[:nl_pos + 1]
                del partial[:nl_pos + 1]

                line = line_bytes.decode("utf-8", errors="replace")
                sys.stdout.write(colorize_line(line))
                sys.stdout.flush()

        if partial:
            line = partial.decode("utf-8", errors="replace")
            sys.stdout.write(colorize_line(line))
            sys.stdout.flush()

    finally:
        ser.close()
        print(f"\n{MAGENTA}Closed {args.port}{RESET}", file=sys.stderr)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())

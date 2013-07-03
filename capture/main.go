package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/akrennmair/gopcap"
)

func parseCommand(msg string) (cmd string, flag bool) {
	flag = false
	switch {
	case strings.HasPrefix(msg, "set"):
		fallthrough
	case strings.HasPrefix(msg, "cas"):
		fallthrough
	case strings.HasPrefix(msg, "add"):
		flag = true
		cmd = strings.Join(strings.SplitN(msg, " ", 6)[:5], " ")

	case strings.HasPrefix(msg, "get"):
		fallthrough
	case strings.HasPrefix(msg, "getl"):
		fallthrough
	case strings.HasPrefix(msg, "delete"):
		flag = true
		cmd = strings.Join(strings.SplitN(msg, " ", 3)[:2], " ")
	}
	return
}

func main() {
	var device *string = flag.String("i", "", "interface")
	var infile *string = flag.String("r", "", "pcap file")
	var snaplen *int = flag.Int("s", 65535, "snaplen")
	var outfile *string = flag.String("o", "outfile.cap", "capture file")
	var h *pcap.Pcap
	var err error
	expr := "tcp dst port 11211"

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [ -i interface | -r file.pcap ] [ -o filename.cap ]\n", os.Args[0])
		os.Exit(1)
	}

	flag.Parse()

	switch {

	case *device != "":
		h, err = pcap.Openlive(*device, int32(*snaplen), true, 0)
	case *infile != "":
		h, err = pcap.Openoffline(*infile)
	default:
		devs, err := pcap.Findalldevs()
		if err != nil || len(devs) == 0 {
			fmt.Fprintf(os.Stderr, "Couldn't find any devices: %s\n", err)
			os.Exit(1)
		}
		h, err = pcap.Openlive(devs[0].Name, int32(*snaplen), true, 0)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize pcap (%s)\n", err)
		os.Exit(1)
	}

	defer h.Close()

	if expr != "" {
		ferr := h.Setfilter(expr)
		if ferr != nil {
			fmt.Fprintf(os.Stderr, "Unable to set filter :%s (%s)\n", expr, ferr)
			os.Exit(1)
		}
	}

	f, ferr := os.Create(*outfile)
	if ferr != nil {
		fmt.Fprintf(os.Stderr, "Unable to create file:%s (%s)\n", *outfile, ferr)
		os.Exit(1)
	}
	defer f.Close()

	stream := bufio.NewWriter(f)
	defer stream.Flush()

	for pkt := h.Next(); ; pkt = h.Next() {
		if pkt == nil {
			continue
		}

		pkt.Decode()
		if len(pkt.Headers) != 0 {
			for _, msg := range strings.Split(string(pkt.Payload), "\r\n") {

				if cmd, ok := parseCommand(msg); ok {
					fmt.Fprintf(stream, "%d,%s:%d,%s\n", pkt.Time.UnixNano(), pkt.IP.SrcAddr(), pkt.TCP.SrcPort, cmd)
					stream.Flush()
				}
			}
		}
	}
}

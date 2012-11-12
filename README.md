Membase Replay
==============

This tool can be used to capture traffic from membase servers and simulate same
traffic to another membase node. This tool also simulates multiple connections.

# Examples

## Capture
    #  ./capture -help
    usage: ./capture [ -i interface | -r file.pcap ] [ -o filename.cap ]

    # ./capture -i eth0 -o data.cap

## Replay
    # ./replay -help
    Usage ./replay: -c data.cap -h server:port -s scale

    # ./replay -c data.cap -h somehost:11211

    # ./replay -c data.cap -h somehost:11211 -s 2  # 2x rate

The binaries for Centos 5.4 is available under bin directory

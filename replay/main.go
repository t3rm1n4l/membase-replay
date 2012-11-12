package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)


var wait sync.WaitGroup
var connections map[string]chan string
var count int

func genRequest(cmd string) string {
	params := strings.Split(cmd, " ")

	var key string
	var flag int
	var exp int
	var length int
	ret := ""

	switch params[0] {
	case "set":
		fallthrough
	case "cas":
		fallthrough
	case "add":
		key = params[1]
		flag, _ = strconv.Atoi(params[2])
		exp, _ = strconv.Atoi(params[3])
		length, _ = strconv.Atoi(params[4])
		ret = fmt.Sprintf("set %s %d %d %d 0001:\r\n%s\r\n", key, flag, exp, length, strings.Repeat("x", length))

	case "get":
		fallthrough
	case "getl":
		key = params[1]
		ret = fmt.Sprintf("get %s\r\n", key)

	case "delete":
		key = params[1]
		ret = fmt.Sprintf("delete %s\n", key)
	}
	return ret

}

func respReader(conn *net.Conn) {
	reader := bufio.NewReader(*conn)
	defer wait.Done()

	for {
		_, err := reader.ReadString('\n')
		if err != nil {
			return
		}

	}
}

func handleConnection(server string, ch chan string) {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect server:%s (%s)\n", server, err)
		os.Exit(1)
	}
	defer conn.Close()
	defer wait.Done()

	wait.Add(1)
	go respReader(&conn)

	for {
		msg, ok := <-ch
		if !ok {
			return
		}
		conn.Write([]byte(msg))
	}

}

func main() {
	var capfile *string = flag.String("c", "", "Capture data file")
	var server *string = flag.String("h", "127.0.0.1:11211", "Server address")
	connections = make(map[string]chan string)
	count = 0

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage %s: -c data.cap -h server:port\n", os.Args[0])
		os.Exit(1)
	}

	flag.Parse()

	f, err := os.Open(*capfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open file:%s (%s)\n", *capfile, err)
		os.Exit(1)
	}
	defer f.Close()
	reader := csv.NewReader(f)
	replayStartTime := time.Now()
	var actualStartTime time.Time
	firstEntry := false

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		v, _ := strconv.ParseUint(record[0],10,64)
		recordTime := time.Unix(0, int64(v))

		if !firstEntry {
			firstEntry = true
			actualStartTime = recordTime
		}

		toSleep := recordTime.Sub(actualStartTime) - time.Now().Sub(replayStartTime)
		time.Sleep(toSleep)
		ch := connections[record[1]]
		if ch == nil {
			ch = make(chan string)
			connections[record[1]] = ch
			wait.Add(1)
			go handleConnection(*server, ch)
		}
		ch <- genRequest(record[2])
	}

	for _, v := range connections {
		close(v)
	}
	wait.Wait()

}

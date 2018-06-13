package main

import (
	"flag"
	"log"
	"net"
	"strconv"
	"time"
	"sync"
)

const (
	SENDLEN  = 65535
	TIMEOUT  = time.Second * 12
	MAXSLEEP = 300
	MOD      = 187
)

func reverse(s []byte, l int) {
	for i := 0; i < l; i++ {
		s[i] = s[i] ^ MOD
	}
}

func eachConn(remote string, tc net.Conn) {
	uc, err := net.DialTimeout("tcp", remote, TIMEOUT)
	if err != nil {
		log.Println("get remote conn :", err.Error())
		return
	}

	var pipemux sync.Mutex
	pipe := false

	go netCopy(uc, tc, pipemux, pipe)
	go netCopy(tc, uc, pipemux, pipe)

}

func netCopy(src, dst net.Conn, mux sync.Mutex, pipe bool) {
	defer func() {
		mux.Lock()
		if !pipe {
			if src != nil {
				src.Close()
			}
			if dst != nil {
				dst.Close()
			}
		}
		mux.Unlock()
	}()
	buf := make([]byte, SENDLEN)
	for idx := 0; idx < MAXSLEEP; idx++ {
		src.SetReadDeadline(time.Now().Add(TIMEOUT))
		nr, err := src.Read(buf)
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				continue
			}
			log.Println(src.RemoteAddr(), err.Error())
			break
		}
		if nr > 0 {
			idx = 0
			reverse(buf, nr)
			dst.SetWriteDeadline(time.Now().Add(TIMEOUT))
			_, err = dst.Write(buf[:nr])
			if err != nil {
				log.Println(dst.RemoteAddr(), err.Error())
				break
			}
		}
	}
}

func main() {
	var host string
	var port int64
	var local string
	var pcount int64

	flag.StringVar(&local, "local", ":3000", "local port")
	flag.StringVar(&host, "host", "127.0.0.1", "host")
	flag.Int64Var(&port, "port", 3000, "port")
	flag.Int64Var(&pcount, "pcount", 20, "port count")

	help := flag.Bool("help", false, "Display usage")
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		return
	}

	l, err := net.Listen("tcp", local)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer l.Close()
	var idx int64 = 0
	for {
		tc, err := l.Accept()
		if err != nil {
			log.Println("accept tcp conn :", err.Error())
			continue
		}

		go eachConn(host+":"+strconv.FormatInt(port+idx, 10), tc)
		idx++
		if idx == pcount {
			idx = 0
		}
	}
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

package main

import (
	"flag"
	"log"
	"net"
	"time"
	"strconv"
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
	uc, err := net.Dial("tcp", remote)
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
			return
		}
		if nr > 0 {
			idx = 0
			reverse(buf, nr)
			dst.SetWriteDeadline(time.Now().Add(TIMEOUT))
			_, err = dst.Write(buf[0:nr])
			if err != nil {
				log.Println(dst.RemoteAddr(), err.Error())
				return
			}
		}
	}
}

func eachListen(listen, backend string) {
	l, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer l.Close()

	for {
		tc, err := l.Accept()
		if err != nil {
			log.Println("accept tcp conn :", err.Error())
			continue
		}
		go eachConn(backend, tc)
	}
}

func main() {
	var remote string
	var local int64
	var pcount int64

	flag.Int64Var(&local, "local", 3000, "local port")
	flag.Int64Var(&pcount, "pcount", 20, "port count")
	flag.StringVar(&remote, "remote", "127.0.0.1:3306", "remote")
	help := flag.Bool("help", false, "Display usage")
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		return
	}
	max := local + pcount
	for ; local < max; local++ {
		go eachListen(":"+strconv.FormatInt(local, 10), remote)
	}

	for {
		time.Sleep(time.Minute)
	}
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

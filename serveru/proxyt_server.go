package main

import (
	"flag"
	"log"
	"net"
	"time"
	"strconv"

	"github.com/xtaci/kcp-go"
)

const (
	SENDLEN  = 65535
	TIMEOUT  = time.Second * 12
	MAXSLEEP = 300
)

func reverse(s []byte, l int) {
	for i := 0; i < l; i++ {
		s[i] = byte(uint8(0xff) - uint8(s[i]))
	}
}

func eachConn(remote string, tc net.Conn) {
	uc, err := net.Dial("tcp", remote)
	defer func() {
		if tc != nil {
			tc.Close()
		}
		if uc != nil {
			uc.Close()
		}
	}()
	if err != nil {
		log.Println("get remote conn :", err.Error())
		return
	}
	ch1 := make(chan bool)
	ch2 := make(chan bool)

	go netCopy(uc, tc, ch1)
	go netCopy(tc, uc, ch2)

	select {
	case <-ch1:
		break
	case <-ch2:
		break
	}
}

func netCopy(src, dst net.Conn, ch chan bool) {
	defer func() {
		ch <- true
		close(ch)
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
	l,err := kcp.
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

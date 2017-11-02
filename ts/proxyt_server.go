package main

import (
	"flag"
	"log"
	"net"
	"time"
	"strconv"
)

const (
	SENDLEN = 1280
	TIMEOUT = time.Second*12
)

func reverse(s []byte) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
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
	for {
		src.SetReadDeadline(time.Now().Add(TIMEOUT))
		nr, err := src.Read(buf)
		if err != nil {
			log.Println(src.RemoteAddr(), err.Error())
			return
		}
		tmp_buf:=buf[0:nr]
		reverse(tmp_buf)
		if nr > 0 {
			dst.SetWriteDeadline(time.Now().Add(TIMEOUT))
			_, err = dst.Write(buf[0:nr])
			if err != nil {
				log.Println(dst.RemoteAddr(),err.Error())
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
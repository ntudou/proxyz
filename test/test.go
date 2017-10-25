package main

import (
	"net"
	"log"
)

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
	defer close(ch)
	buf := make([]byte, 1024)
	for {
		nr, err := src.Read(buf)
		if err != nil {
			log.Println(src.RemoteAddr(), err.Error())
			break
		}
		if nr > 0 {
			_, err = dst.Write(buf[0:nr])
			if err != nil {
				log.Println(err.Error())
				break
			}
		}

	}
	ch <- true
}

func main() {
	l, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer l.Close()

	for {
		tc, err := l.Accept()
		if err != nil {
			log.Println("accept tcp conn :", err.Error())
			tc.Close()
			continue
		}
		go eachConn("138.128.203.86:8000", tc)
	}
}

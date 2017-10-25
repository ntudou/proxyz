package main

import (
	"flag"
	"log"
	"net"
	"time"
	"strconv"
)

func eachConn(remote string, tc net.Conn) {
	uc, err := net.DialTimeout("tcp", remote, time.Second*6)
	if err != nil {
		log.Println("get remote conn :", err.Error())
		if uc != nil {
			uc.Close()
		}
		return
	}
	ch1 := make(chan bool)
	ch2 := make(chan bool)
	go netCompress(tc, uc, ch1,ch2)
	go netUnCompress(uc, tc, ch2,ch1)
}

func netCompress(src, dst net.Conn, ch1,ch2 chan bool) error {
	buf := make([]byte, 1024)
	var err error
	for {
		select {
		case <-ch1:
			break
		default:
			src.SetReadDeadline(time.Now().Add(time.Second))
			nr, err := src.Read(buf)
			if err != nil {
				if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
					continue
				}
				log.Println(src.RemoteAddr(),err.Error())
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
	}
	ch2 <- true
	close(ch1)
	return err
}
func netUnCompress(src, dst net.Conn, ch1,ch2 chan bool) error {
	buf := make([]byte, 1024)
	var err error
	for {
		select {
		case <-ch1:
			break
		default:
			src.SetReadDeadline(time.Now().Add(time.Second))
			nr, err := src.Read(buf)
			if err != nil {
				if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
					continue
				}
				log.Println(err.Error())
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
	}
	ch2 <- true
	close(ch1)
	return err
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
			tc.Close()
			continue
		}

		go eachConn(host+":"+strconv.FormatInt(port+idx, 10), tc)
		idx++
		if idx == port+pcount {
			idx = 0
		}
	}

	for {
		time.Sleep(time.Minute)
	}
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

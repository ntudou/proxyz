package main

import (
	"flag"
	"log"
	"net"
	"time"
	"strconv"
	"bytes"
	"compress/zlib"
	"io"
)

const (
	VCODE=3
	SENDLEN=51200
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
	buf := make([]byte, SENDLEN)
	for {
		nr, err := src.Read(buf)
		if err != nil {
			log.Println(src.RemoteAddr(), err.Error())
			break
		}
		if nr > 0 {
			_, err = dst.Write(buf[0:nr])
			if err != nil {
				log.Println(dst.RemoteAddr(),err.Error())
				break
			}
		}
	}
	ch <- true
}

func netCompress(src, dst net.Conn,ch chan bool) {
	defer close(ch)
	for {
		buf := make([]byte, VCODE)
		nr, err := src.Read(buf)
		if err != nil {
			log.Println(src.RemoteAddr(),err.Error())
			break
		}
		if nr!=VCODE{
			log.Println(src.RemoteAddr(),"vcode length err")
			break
		}else {


			var in bytes.Buffer
			w := zlib.NewWriter(&in)
			w.Write(buf[0:nr])
			w.Close()
			_, err = dst.Write(in.Bytes())
			if err != nil {
				log.Println(err.Error())
				break
			}
		}
	}
	ch<-true
}
func netUnCompress(src, dst net.Conn) error {
	buf := make([]byte, 1024)
	var err error
	for {
		nr, err := src.Read(buf)
		if err != nil {
			log.Println(err.Error())
			break
		}
		if nr > 0 {
			rb := bytes.NewReader(buf[0:nr])
			out := &bytes.Buffer{}
			r, err := zlib.NewReader(rb)
			if r==nil{
				log.Println(err.Error())
				break
			}
			io.Copy(out, r)
			r.Close()
			_, err = dst.Write(out.Bytes())
			if err != nil {
				log.Println(err.Error())
				break
			}
		}
	}
	return err
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
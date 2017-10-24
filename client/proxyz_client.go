package client

import (
	"flag"
	"log"
	"net"
	"time"
	"strconv"
	"compress/zlib"
	"bytes"
	"io"
)

func eachConn(remote string, tc net.Conn) {
	uc, err := net.DialTimeout("tcp", remote,time.Minute)
	if err != nil {
		log.Println("get remote conn :", err.Error())
		uc.Close()
		return
	}
	go netCompress(uc,tc)
	go netUnCompress(tc,uc)
}

func netCompress(src, dst net.Conn) error {
	buf := make([]byte, 1024)
	var err error
	for {
		nr, err := src.Read(buf)
		if err!=nil{
			break
		}
		if nr > 0 {
			var in *bytes.Buffer
			w := zlib.NewWriter(in)
			w.Write(buf[0:nr])
			w.Close()
			_, err = dst.Write(in.Bytes())
			if err != nil {
				break
			}
		}
	}
	return err
}
func netUnCompress(src, dst net.Conn) error {
	buf := make([]byte, 1024)
	var err error
	for {
		nr, err := src.Read(buf)
		if err != nil {
			break
		}
		if nr > 0 {
			rb := bytes.NewReader(buf[0:nr])
			var out *bytes.Buffer
			r, err := zlib.NewReader(rb)
			if err != nil {
				break
			}
			io.Copy(out, r)
			r.Close()
			_, err = dst.Write(out.Bytes())
			if err != nil {
				break
			}
		}
	}
	return err
}

func eachListen(listen, backend string) error {
	l, err := net.Listen("tcp", listen)
	if err != nil {
		return err
	}
	defer l.Close()

	for {
		tc, err := l.Accept()
		if err != nil {
			log.Println("accept tcp conn :", err.Error())
			tc.Close()
			continue
		}

		go eachConn(backend, tc)
	}
	return nil
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
	max := port + pcount
	for ; port < max; port++ {
		err := eachListen(local,":"+strconv.FormatInt(port, 10))
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	for {
		time.Sleep(time.Minute)
	}
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

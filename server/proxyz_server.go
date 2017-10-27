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
	"encoding/binary"
)

const (
	XCODE   = 0x00
	VCODE   = 1
	VLENGTH = 2
	SENDLEN = 1280
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

	go netCompress(uc,tc,ch1)
	go netUnCompress(tc,uc,ch2)

	select {
	case <-ch1:
		break
	case <-ch2:
		break
	}
}

func netUnCompress(src, dst net.Conn, ch chan bool) {
	defer func() {
		ch <- true
		close(ch)
	}()
	for {
		code_buf := make([]byte, VCODE)
		nr, err := src.Read(code_buf)
		if err != nil {
			log.Println(src.RemoteAddr(), err.Error())
			return
		}
		if XCODE != code_buf[0] {
			continue
		}
		len_buf := make([]byte, VLENGTH)
		nr, err = src.Read(len_buf)
		if err != nil {
			log.Println(src.RemoteAddr(), err.Error())
			return
		}
		if nr != 2 {
			continue
		}
		body_len := binary.BigEndian.Uint16(len_buf)
		if body_len > SENDLEN {
			continue
		}

		var body_buf []byte
		for{
			x_buf := make([]byte, body_len)
			nr, err = src.Read(x_buf)
			if err!=nil{
				log.Print(src.RemoteAddr(),err.Error())
				return
			}
			if nr>0{
				body_buf=append(body_buf,x_buf...)
			}
			body_len=body_len-uint16(nr)
			if body_len == 0 {
				err=unCompress(dst,body_buf)
				if err !=nil{
					log.Print(dst.RemoteAddr(),err.Error())
					return
				}
				break
			}
		}
	}
}

func netCompress(src, dst net.Conn, ch chan bool) {
	defer func() {
		ch <- true
		close(ch)
	}()
	buf := make([]byte, SENDLEN)
	for {
		nr, err := src.Read(buf)
		if err != nil {
			log.Println(src.RemoteAddr(), err.Error())
			return
		}
		if nr > 0 {
			err = compress(dst,buf[0:nr])
			if err != nil {
				log.Println(dst.RemoteAddr(), err.Error())
				return
			}
		}
	}
}

func compress(dst net.Conn, buf []byte) error {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	w.Write(buf)
	w.Close()
	var packet_buf []byte
	var body_len_buf []byte
	binary.BigEndian.PutUint16(body_len_buf, uint16(in.Len()))
	packet_buf = append(packet_buf, XCODE)
	packet_buf = append(packet_buf, body_len_buf...)
	packet_buf = append(packet_buf, in.Bytes()...)
	_, err := dst.Write(packet_buf)
	if err != nil {
		return  err
	}
	return nil
}

func unCompress(dst net.Conn, buf []byte) error {
	rb := bytes.NewReader(buf)
	out := &bytes.Buffer{}
	r, err := zlib.NewReader(rb)
	if err!= nil {
		return err
	}
	io.Copy(out, r)
	r.Close()
	_, err = dst.Write(out.Bytes())
	if err != nil {
		return err
	}
	return  nil
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

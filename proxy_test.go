package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"
)

func TestSocks5ConnectRequestDomain(t *testing.T) {
	req, err := buildSocks5ConnectRequest("example.com", "3389")
	if err != nil {
		t.Fatal(err)
	}
	wantPrefix := []byte{0x05, 0x01, 0x00, 0x03, byte(len("example.com"))}
	if !bytes.HasPrefix(req, append(wantPrefix, []byte("example.com")...)) {
		t.Fatalf("unexpected request prefix: %v", req)
	}
	port := binary.BigEndian.Uint16(req[len(req)-2:])
	if port != 3389 {
		t.Fatalf("unexpected port: %d", port)
	}
}

func TestProxyServerThroughLocalSocks5(t *testing.T) {
	setCurrentLanguage("en-US")
	targetLn := listenLocal(t)
	defer targetLn.Close()
	go func() {
		conn, err := targetLn.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		line, _ := bufio.NewReader(conn).ReadString('\n')
		_, _ = conn.Write([]byte("echo:" + line))
	}()

	socksLn := listenLocal(t)
	defer socksLn.Close()
	go serveOneSocks5(t, socksLn)

	var srv ProxyServer
	listenLn := listenLocal(t)
	listenAddr := listenLn.Addr().(*net.TCPAddr)
	listenLn.Close()

	targetAddr := targetLn.Addr().(*net.TCPAddr)
	socksAddr := socksLn.Addr().(*net.TCPAddr)
	err := srv.Start(ProxyConfig{
		ListenHost: "127.0.0.1",
		ListenPort: itoa(listenAddr.Port),
		TargetHost: "127.0.0.1",
		TargetPort: itoa(targetAddr.Port),
		SocksHost:  "127.0.0.1",
		SocksPort:  itoa(socksAddr.Port),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Stop()

	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", itoa(listenAddr.Port)), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_, _ = conn.Write([]byte("hello\n"))
	got, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if got != "echo:hello\n" {
		t.Fatalf("got %q", got)
	}
}

func listenLocal(t *testing.T) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	return ln
}

func serveOneSocks5(t *testing.T, ln net.Listener) {
	conn, err := ln.Accept()
	if err != nil {
		return
	}
	defer conn.Close()
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		t.Error(err)
		return
	}
	methods := make([]byte, int(header[1]))
	if _, err := io.ReadFull(conn, methods); err != nil {
		t.Error(err)
		return
	}
	_, _ = conn.Write([]byte{0x05, 0x00})

	reqHeader := make([]byte, 4)
	if _, err := io.ReadFull(conn, reqHeader); err != nil {
		t.Error(err)
		return
	}
	if reqHeader[1] != 0x01 {
		t.Errorf("unexpected command: %d", reqHeader[1])
		return
	}
	host := ""
	switch reqHeader[3] {
	case 0x01:
		ip := make([]byte, 4)
		_, _ = io.ReadFull(conn, ip)
		host = net.IP(ip).String()
	case 0x03:
		l := make([]byte, 1)
		_, _ = io.ReadFull(conn, l)
		name := make([]byte, int(l[0]))
		_, _ = io.ReadFull(conn, name)
		host = string(name)
	default:
		t.Errorf("unsupported address type: %d", reqHeader[3])
		return
	}
	portBuf := make([]byte, 2)
	_, _ = io.ReadFull(conn, portBuf)
	port := int(binary.BigEndian.Uint16(portBuf))
	target, err := net.DialTimeout("tcp", net.JoinHostPort(host, itoa(port)), time.Second)
	if err != nil {
		t.Error(err)
		return
	}
	defer target.Close()
	_, _ = conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 127, 0, 0, 1, 0, 0})
	done := make(chan struct{}, 2)
	go func() {
		_, _ = io.Copy(target, conn)
		done <- struct{}{}
	}()
	go func() {
		_, _ = io.Copy(conn, target)
		done <- struct{}{}
	}()
	<-done
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	return string(b[i:])
}

package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ProxyConfig struct {
	ListenHost string
	ListenPort string
	TargetHost string
	TargetPort string
	SocksHost  string
	SocksPort  string
	Username   string
	Password   string
}

type ProxyServer struct {
	mu       sync.Mutex
	cfg      ProxyConfig
	listener net.Listener
	cancel   context.CancelFunc
	running  bool
	lastErr  error
	onError  func(error)
}

func (s *ProxyServer) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *ProxyServer) Start(cfg ProxyConfig) error {
	if err := validateProxyConfig(cfg); err != nil {
		return err
	}
	listenAddr := net.JoinHostPort(cfg.ListenHost, cfg.ListenPort)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("%s: %w", tr(currentLanguage()).ListenFailed, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		_ = ln.Close()
		cancel()
		return nil
	}
	s.cfg = cfg
	s.listener = ln
	s.cancel = cancel
	s.running = true
	s.lastErr = nil
	s.mu.Unlock()

	go s.acceptLoop(ctx, ln)
	return nil
}

func (s *ProxyServer) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	ln := s.listener
	cancel := s.cancel
	s.running = false
	s.listener = nil
	s.cancel = nil
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if ln != nil {
		return ln.Close()
	}
	return nil
}

func (s *ProxyServer) acceptLoop(ctx context.Context, ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				s.reportError(err)
				return
			}
		}
		go s.handleConn(ctx, conn)
	}
}

func (s *ProxyServer) handleConn(ctx context.Context, client net.Conn) {
	defer client.Close()

	s.mu.Lock()
	cfg := s.cfg
	s.mu.Unlock()

	socksConn, err := net.DialTimeout("tcp", net.JoinHostPort(cfg.SocksHost, cfg.SocksPort), 15*time.Second)
	if err != nil {
		s.reportError(fmt.Errorf("%s: %w", tr(currentLanguage()).SocksFailed, err))
		return
	}
	defer socksConn.Close()

	if err := socks5Connect(socksConn, cfg.TargetHost, cfg.TargetPort, cfg.Username, cfg.Password); err != nil {
		s.reportError(err)
		return
	}

	done := make(chan struct{}, 2)
	copyConn := func(dst, src net.Conn) {
		_, _ = io.Copy(dst, src)
		_ = dst.SetDeadline(time.Now())
		_ = src.SetDeadline(time.Now())
		done <- struct{}{}
	}
	go copyConn(socksConn, client)
	go copyConn(client, socksConn)

	select {
	case <-ctx.Done():
	case <-done:
	}
}

func (s *ProxyServer) reportError(err error) {
	s.mu.Lock()
	s.lastErr = err
	cb := s.onError
	s.mu.Unlock()
	if cb != nil {
		cb(err)
	}
}

func validateProxyConfig(cfg ProxyConfig) error {
	if strings.TrimSpace(cfg.ListenHost) == "" || strings.TrimSpace(cfg.ListenPort) == "" {
		return errors.New(tr(currentLanguage()).NeedListen)
	}
	if strings.TrimSpace(cfg.TargetHost) == "" || strings.TrimSpace(cfg.TargetPort) == "" {
		return errors.New(tr(currentLanguage()).NeedTarget)
	}
	if strings.TrimSpace(cfg.SocksHost) == "" || strings.TrimSpace(cfg.SocksPort) == "" {
		return errors.New(tr(currentLanguage()).NeedSocks)
	}
	for _, p := range []string{cfg.ListenPort, cfg.TargetPort, cfg.SocksPort} {
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 || n > 65535 {
			return fmt.Errorf("invalid port: %s", p)
		}
	}
	return nil
}

func socks5Connect(conn net.Conn, host, port, username, password string) error {
	methods := []byte{0x00}
	if username != "" || password != "" {
		methods = append(methods, 0x02)
	}
	if _, err := conn.Write(append([]byte{0x05, byte(len(methods))}, methods...)); err != nil {
		return err
	}
	reply := make([]byte, 2)
	if _, err := io.ReadFull(conn, reply); err != nil {
		return err
	}
	if reply[0] != 0x05 {
		return errors.New(tr(currentLanguage()).SocksRejected)
	}
	switch reply[1] {
	case 0x00:
	case 0x02:
		if err := socks5Auth(conn, username, password); err != nil {
			return err
		}
	case 0xff:
		return errors.New(tr(currentLanguage()).UnsupportedAuth)
	default:
		return fmt.Errorf("%s: method %d", tr(currentLanguage()).UnsupportedAuth, reply[1])
	}

	req, err := buildSocks5ConnectRequest(host, port)
	if err != nil {
		return err
	}
	if _, err := conn.Write(req); err != nil {
		return err
	}
	return readSocks5ConnectReply(conn)
}

func socks5Auth(conn net.Conn, username, password string) error {
	if len(username) > 255 || len(password) > 255 {
		return errors.New("SOCKS5 username or password is too long")
	}
	req := []byte{0x01, byte(len(username))}
	req = append(req, []byte(username)...)
	req = append(req, byte(len(password)))
	req = append(req, []byte(password)...)
	if _, err := conn.Write(req); err != nil {
		return err
	}
	reply := make([]byte, 2)
	if _, err := io.ReadFull(conn, reply); err != nil {
		return err
	}
	if reply[0] != 0x01 || reply[1] != 0x00 {
		return errors.New(tr(currentLanguage()).AuthRejected)
	}
	return nil
}

func buildSocks5ConnectRequest(host, port string) ([]byte, error) {
	portNum, err := strconv.Atoi(port)
	if err != nil || portNum < 1 || portNum > 65535 {
		return nil, fmt.Errorf("invalid port: %s", port)
	}
	req := []byte{0x05, 0x01, 0x00}
	if ip := net.ParseIP(host); ip != nil {
		if v4 := ip.To4(); v4 != nil {
			req = append(req, 0x01)
			req = append(req, v4...)
		} else {
			req = append(req, 0x04)
			req = append(req, ip.To16()...)
		}
	} else {
		if len(host) > 255 {
			return nil, errors.New("target host is too long")
		}
		req = append(req, 0x03, byte(len(host)))
		req = append(req, []byte(host)...)
	}
	var p [2]byte
	binary.BigEndian.PutUint16(p[:], uint16(portNum))
	req = append(req, p[:]...)
	return req, nil
}

func readSocks5ConnectReply(conn net.Conn) error {
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return err
	}
	if header[0] != 0x05 || header[1] != 0x00 {
		return fmt.Errorf("%s: code %d", tr(currentLanguage()).SocksRejected, header[1])
	}
	var addrLen int
	switch header[3] {
	case 0x01:
		addrLen = 4
	case 0x03:
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			return err
		}
		addrLen = int(lenBuf[0])
	case 0x04:
		addrLen = 16
	default:
		return errors.New(tr(currentLanguage()).SocksRejected)
	}
	if addrLen > 0 {
		if _, err := io.CopyN(io.Discard, conn, int64(addrLen)); err != nil {
			return err
		}
	}
	if _, err := io.CopyN(io.Discard, conn, 2); err != nil {
		return err
	}
	return nil
}

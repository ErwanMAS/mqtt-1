package server

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/2hdddg/mqtt/packet"
)

type ReaderFake struct {
	pack chan packet.Packet
	err  chan error
}

func tNewReaderFake(t *testing.T) *ReaderFake {
	return &ReaderFake{
		pack: make(chan packet.Packet, 1),
		err:  make(chan error, 1),
	}
}

// Implements Reader interface
func (r *ReaderFake) ReadPacket(version uint8) (packet.Packet, error) {
	for {
		select {
		case p := <-r.pack:
			return p, nil
		case e := <-r.err:
			return nil, e
		}
	}
}

func (r *ReaderFake) tWritePacket(p packet.Packet) {
	r.pack <- p
}

type WriterFake struct {
	err     error
	written chan packet.Packet
}

func (w *WriterFake) WritePacket(p packet.Packet) error {
	w.written <- p
	return w.err
}

type ConnFake struct {
	closed bool
}

func (c *ConnFake) Read(b []byte) (n int, err error) {
	return 0, nil
}
func (c *ConnFake) Write(b []byte) (n int, err error) {
	return 0, nil
}
func (c *ConnFake) Close() error {
	c.closed = true
	return nil
}
func (c *ConnFake) LocalAddr() net.Addr {
	return nil
}
func (c *ConnFake) RemoteAddr() net.Addr {
	return nil
}
func (c *ConnFake) SetDeadline(t time.Time) error {
	return nil
}
func (c *ConnFake) SetReadDeadline(t time.Time) error {
	return nil
}
func (c *ConnFake) SetWriteDeadline(t time.Time) error {
	return nil
}

type AuthFake struct {
}

func (a *AuthFake) CheckConnect(c *packet.Connect) packet.ConnRetCode {
	return packet.ConnAccepted
}

type PubFake struct {
	publishChan chan *packet.Publish
}

func NewPubFake() *PubFake {
	return &PubFake{
		publishChan: make(chan *packet.Publish, 1),
	}
}

func (f *PubFake) Publish(s *Session, p *packet.Publish) error {
	f.publishChan <- p
	return nil
}

type tLogger struct {
}

func (l *tLogger) Info(s string) {
	fmt.Println(s)
}

func (l *tLogger) Error(s string) {
	fmt.Println(s)
}

func (l *tLogger) Debug(s string) {
	fmt.Println(s)
}

func tSession(
	t *testing.T) (*Session, *ReaderFake, *WriterFake, *PubFake) {

	rd := tNewReaderFake(t)
	connect := &packet.Connect{
		ProtocolName:     "MQTT",
		ProtocolVersion:  4,
		KeepAliveSecs:    30,
		ClientIdentifier: "xyz",
	}
	wr := &WriterFake{
		written: make(chan packet.Packet, 3),
	}
	conn := &ConnFake{}
	pub := NewPubFake()
	sess := newSession(conn, rd, wr, connect)
	sess.Start(pub, &tLogger{})
	return sess, rd, wr, pub
}


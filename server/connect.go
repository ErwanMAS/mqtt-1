package server

import (
	"errors"

	"net"
	"time"

	packet "github.com/2hdddg/mqtt/controlpacket"
)

func Connect(conn net.Conn, r Reader, w Writer) (*Session, error) {
	// If server does not receive CONNECT in a reasonable amount of time,
	// the server should close the network connection.
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Wait for CONNECT
	p, err := r.ReadPacket(5)
	if err != nil {
		return nil, err
	}

	c, ok := p.(*packet.Connect)
	if !ok {
		conn.Close()
		return nil, errors.New("Wrong package")
	}

	// If this fails, the server should close the connection without
	// sending a CONNACK
	if c.ProtocolName != "MQTT" && c.ProtocolName != "MQIsdp" {
		conn.Close()
		return nil, errors.New("Invalid protocol")
	}

	// If version check fails, notify client about wrong version
	if c.ProtocolVersion < 4 || c.ProtocolVersion > 5 {
		w.WriteAckConnection(
			packet.RefuseConnection(packet.ConnRefusedVersion))
		conn.Close()
		return nil, errors.New("Protocol version not supported")
	}

	// Accept connection by acking it
	ack := &packet.AckConnection{
		SessionPresent: false, // TODO:
		RetCode:        packet.ConnAccepted,
	}
	err = w.WriteAckConnection(ack)
	if err != nil {
		conn.Close()
		return nil, errors.New("Failed to send CONNACK")
	}

	// TODO: Hook

	// Reset deadline after CONNECT received
	conn.SetReadDeadline(time.Time{})

	return &Session{
		conn:       conn,
		rd:         r,
		wr:         w,
		connPacket: c,
	}, nil
}

package packet

type Publish struct {
	Duplicate bool
	QoS       QoS
	Retain    bool
	Topic     string
	PacketId  uint16
}

func (r *Reader) readPublish(fixflags uint8) (*Publish, error) {
	const C = "read PUBLISH"
	p := &Publish{}

	// From fixed header
	p.Duplicate = (fixflags & 0x08) > 0
	p.Retain = (fixflags & 0x01) > 0
	qoS := (fixflags >> 1) & 0x03
	if qoS == 4 {
		// A PUBLISH Packet MUST NOT have both QoS bits set to 1.
		// If a Server or Client receives a PUBLISH Packet which has
		// both QoS bits set to 1 it MUST close the Network Connection
		return nil, &Error{c: C, m: "QoS illegal"}
	}
	p.QoS = QoS(qoS)

	// From variable header
	var err error
	p.Topic, err = r.str()
	if err != nil {
		return nil, &Error{c: C, m: "Topic name", err: err}
	}

	if qoS > 0 {
		p.PacketId, err = r.int2()
		if err != nil {
			return nil, &Error{c: C, m: "Packet identifier", err: err}
		}
	}

	return p, nil
}

func (p *Publish) toPacket() []byte {
	// Variable header
	v := strToBytes(p.Topic)
	if p.QoS == QoS1 || p.QoS == QoS2 {
		v = append(v, toInt2(p.PacketId)...)
	}

	rem := toVarInt(uint32(len(v)))
	h := make([]uint8, 1, len(v)+len(rem)+1)
	flags := flagsToBitsU8([]bool{
		p.Duplicate,
		false,
		false,
		p.Retain,
	})
	h[0] = uint8(PUBLISH<<4) | (flags | (uint8(p.QoS) << 1))
	h = append(h, rem...)
	h = append(h, v...)

	return h
}
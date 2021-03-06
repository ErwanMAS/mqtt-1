package packet

type Subscription struct {
	Topic string
	QoS   QoS
}

type Subscribe struct {
	PacketId      uint16
	Subscriptions []Subscription
}

func (r *Reader) readSubscribe(fixflags uint8) (*Subscribe, error) {
	sub := &Subscribe{}
	var err error

	// From variable header
	sub.PacketId, err = r.int2()
	if err != nil {
		return nil, err
	}

	// Payload, N number of subscriptions
	for {
		s := Subscription{}
		topic, err := r.strMaybe()
		if err != nil {
			return nil, err
		}
		if topic == nil {
			break
		}
		s.Topic = *topic

		qoS, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		if qoS > byte(QoSHighest) {
			return nil, newProtoErr("Illegal QoS")
		}
		s.QoS = QoS(qoS)

		sub.Subscriptions = append(sub.Subscriptions, s)
	}

	// TODO: Check fixflags for correct value, must be 2
	// Remove length for packetid
	//length -= 2
	return sub, nil
}

func (s *Subscribe) toPacket() []byte {
	v := toInt2(s.PacketId)

	// Payload, N number of subscriptions
	for _, x := range s.Subscriptions {
		v = append(v, strToBytes(x.Topic)...)
		v = append(v, byte(x.QoS))
	}

	rem := toVarInt(uint32(len(v)))
	h := make([]uint8, 1, len(v)+len(rem)+1)
	h[0] = uint8(SUBSCRIBE<<4) | 2
	h = append(h, rem...)
	h = append(h, v...)

	return h
}

func (c *Subscribe) name() string {
	return "SUBSCRIBE"
}

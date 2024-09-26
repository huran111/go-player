package rtc

import (
	"encoding/binary"
	"player/internal/av"
	"sync"

	"k8s.io/klog"
)

const (
	RtcPlayerGroup    = "RtcPlayerGroup"
	headerLengthField = 4
)

type Player struct {
	Queue chan *av.Packet
	RtcS  sync.Map
}

func NewPlayer() *Player {

	player := &Player{
		Queue: make(chan *av.Packet, 1024),
	}

	go player.SendPacket()

	return player
}

func (p *Player) AddRtc(rtc *Rtc) error {

	p.RtcS.Store(rtc, rtc)
	return nil

}

func (P *Player) WritePacket(packet *av.Packet) error {

	select {
	case P.Queue <- packet:
	default:
		klog.Info("rtc Que full")
	}
	return nil
}

func (P *Player) VideoPacket(packet *av.Packet) (*rtcPacket, error) {

	outBuf := []byte{}

	for offset := 0; offset < len(packet.Payload); {
		bufferLength := int(binary.BigEndian.Uint32(packet.Payload[offset : offset+headerLengthField]))
		if offset+bufferLength >= len(packet.Payload) {
			break
		}
		offset += headerLengthField
		outBuf = append(outBuf, []byte{0x00, 0x00, 0x00, 0x01}...)
		outBuf = append(outBuf, packet.Payload[offset:offset+bufferLength]...)
		offset += int(bufferLength)
	}

	return &rtcPacket{
		Type:      packet.Tag,
		timestamp: packet.TimeStamp,
		data:      outBuf,
	}, nil
}

func (P *Player) AudioPacket(packet *av.Packet) (*rtcPacket, error) {

	return &rtcPacket{
		Type:      packet.Tag,
		timestamp: packet.TimeStamp,
		data:      packet.Payload,
	}, nil

}

func (p *Player) SendPacket() {

	for q := range p.Queue {

		var err error
		var packet *rtcPacket

		if q.IsAudioPacket() {

			packet, err = p.AudioPacket(q)
			if err != nil {
				klog.Error(err)
			}

		} else {

			packet, err = p.VideoPacket(q)
			if err != nil {
				klog.Error(err)
			}
		}

		p.RtcS.Range(func(key, p interface{}) bool {
			p.(*Rtc).WritPacket(packet)
			return true
		})

	}
}

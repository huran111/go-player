package flv

import "player/internal/av"

type Flv struct {
}

func (f *Flv) WriteVideoPacket(p *av.Packet) error {
	return nil
}

func (f *Flv) WriteAudioPacket(p *av.Packet) error {
	return nil
}

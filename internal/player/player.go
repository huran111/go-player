package player

import "player/internal/av"

type Player interface {
	WritePacket(p *av.Packet) error
}

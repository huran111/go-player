package rtmp

import (
	"errors"
	"player/internal/av"
	"player/internal/player"
	"sync"
)

const (
	maxQueueSize = 1024
)

type Channel struct {
	live        string
	action      chan bool
	players     sync.Map
	packetQueue chan *av.Packet
}

func NewChannel(name string) *Channel {
	w := &Channel{
		live:        name,
		action:      make(chan bool),
		players:     sync.Map{},
		packetQueue: make(chan *av.Packet, maxQueueSize),
	}
	return w
}

func (c *Channel) Name() string {
	return c.live
}

func (c *Channel) LoadPlayer(k string) (player.Player, error) {

	p, ok := c.players.Load(k)
	if !ok {
		return nil, errors.New("player empty")
	}

	return p.(player.Player), nil
}

func (c *Channel) AddPlayer(k string, p player.Player) error {

	_, ok := c.players.Load(k)
	if ok {
		return errors.New("player exist")
	}

	c.players.Store(k, p)

	return nil
}

func (c *Channel) Close() {
	select {
	case c.action <- true:
	default:
	}
}

func (c *Channel) ReadPacket(packet *av.Packet) error {

	select {
	case c.packetQueue <- packet:
	default:
	}
	return nil

}

func (c *Channel) WritePacket() {
	for {
		select {
		case q := <-c.packetQueue:
			c.players.Range(func(key, p interface{}) bool {
				p.(player.Player).WritePacket(q)
				return true
			})

		case <-c.action:
			return
		}
	}
}

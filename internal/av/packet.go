package av

const (
	TAG_AUDIO = 0x08
	TAG_VIDEO = 0x09
)

type Packet struct {
	Tag       int
	TimeStamp uint32 //dts
	Payload   []byte
}

func (p *Packet) IsVideoPacket() bool {
	return p.Tag == TAG_VIDEO
}

func (p *Packet) IsAudioPacket() bool {
	return p.Tag == TAG_AUDIO
}

func NewPacket(Tag int, payload []byte, TimeStamp uint32) (*Packet, error) {

	Packet := &Packet{
		Tag:       Tag,
		Payload:   payload,
		TimeStamp: TimeStamp,
	}
	return Packet, nil
}

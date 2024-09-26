package rtc

import (
	"player/internal/av"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

type Rtc struct {
	*webrtc.PeerConnection
	audioTrack *webrtc.TrackLocalStaticSample
	videoTrack *webrtc.TrackLocalStaticSample
}

type rtcPacket struct {
	Type      int
	timestamp uint32
	data      []byte
}

func (r *Rtc) WritPacket(p *rtcPacket) error {

	if p.Type == av.TAG_VIDEO {

		return r.videoTrack.WriteSample(media.Sample{
			Data:     p.data,
			Duration: time.Second / 30,
		})
	} else {
		return r.audioTrack.WriteSample(media.Sample{
			Data:     p.data,
			Duration: time.Second / 30,
		})
	}
}

func NewRtcPlayer(offer webrtc.SessionDescription) (*Rtc, error) {

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return nil, err
	}

	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	if err != nil {
		return nil, err
	}
	if _, err = peerConnection.AddTrack(videoTrack); err != nil {
		return nil, err
	}

	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypePCMA}, "audio", "pion")
	if err != nil {
		return nil, err
	}
	if _, err = peerConnection.AddTrack(audioTrack); err != nil {
		return nil, err
	}

	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		return nil, err
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, err
	} else if err = peerConnection.SetLocalDescription(answer); err != nil {
		return nil, err
	}
	<-gatherComplete

	return &Rtc{
		PeerConnection: peerConnection,
		audioTrack:     audioTrack,
		videoTrack:     videoTrack,
	}, nil

}

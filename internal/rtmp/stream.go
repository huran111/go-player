package rtmp

import (
	"bytes"
	"io"
	"player/internal/av"

	flv "github.com/yutopp/go-flv/tag"
	"github.com/yutopp/go-rtmp"
	"github.com/yutopp/go-rtmp/message"
	"k8s.io/klog"
)

type Stream struct {
	*Channel
	Chan chan *Channel
	rtmp.DefaultHandler
}

func (s *Stream) OnCreateStream(timestamp uint32, cmd *message.NetConnectionCreateStream) error {
	return nil
}

func (s *Stream) OnPublish(ctx *rtmp.StreamContext, timestamp uint32, cmd *message.NetStreamPublish) error {
	return nil
}

func (s *Stream) OnClose() {
	s.Channel.Close()
}

func (s *Stream) OnConnect(timestamp uint32, cmd *message.NetConnectionConnect) error {

	klog.Info("OnConnect Channel ", cmd.Command.App)

	/*Create a new channel to obtain the ingest stream data,
	  and start a distribution thread to distribute the read and ingest stream data*/
	s.Channel = NewChannel(cmd.Command.App)
	go func() {
		s.Channel.WritePacket()
	}()

	/*notify  Server that a new ingest channel has been added*/
	select {
	case s.Chan <- s.Channel:
	default:
	}

	return nil
}

func (s *Stream) OnAudio(timestamp uint32, payload io.Reader) error {

	var audio flv.AudioData

	if err := flv.DecodeAudioData(payload, &audio); err != nil {
		return err
	}

	data := new(bytes.Buffer)
	if _, err := io.Copy(data, audio.Data); err != nil {
		return err
	}

	packet, err := av.NewPacket(av.TAG_AUDIO, data.Bytes(), timestamp)
	if err != nil {
		klog.Error(err)
		return err
	}

	err = s.Channel.ReadPacket(packet)
	if err != nil {
		klog.Error(err)
		return err
	}

	return nil
}

func (s *Stream) OnVideo(timestamp uint32, payload io.Reader) error {

	var video flv.VideoData
	if err := flv.DecodeVideoData(payload, &video); err != nil {
		return err
	}

	data := new(bytes.Buffer)
	if _, err := io.Copy(data, video.Data); err != nil {
		return err
	}

	packet, err := av.NewPacket(av.TAG_VIDEO, data.Bytes(), timestamp)
	if err != nil {
		klog.Error(err)
		return err
	}

	err = s.Channel.ReadPacket(packet)
	if err != nil {
		klog.Error(err)
		return err
	}

	return nil
}

package web

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"player/internal/config"
	"player/internal/rtc"
	"player/internal/rtmp"

	"github.com/gin-gonic/gin"
	"github.com/pion/webrtc/v3"
	"k8s.io/klog"
)

type CreatePeerConnectRequest struct {
	Url   string `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	Offer webrtc.SessionDescription
}

type channelResponse struct {
	Channel []string `protobuf:"bytes,1,opt,name=Channel,proto3" json:"Channel,omitempty"`
}

type Group struct {
	cfg          *config.Config
	gin          *gin.Engine
	rtmp         *rtmp.Server
	webRtcPlayer *rtc.Player
}

func NewGroup(ctx context.Context, S *rtmp.Server, cfg *config.Config) *Group {

	gin.SetMode(gin.ReleaseMode)

	return &Group{
		gin:  gin.Default(),
		cfg:  cfg,
		rtmp: S,
	}

}

func (s *Group) Run(ctx context.Context) {

	s.gin.LoadHTMLGlob("index/*")
	s.gin.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Main website",
		})
	})

	s.gin.GET("/channels", s.HandleChannels)
	s.gin.POST("/createPeerConnection", s.HandleCreatePeerConn)

	/*listen  http Port*/
	listener, err := net.Listen("tcp", s.cfg.HttpServerAddr)
	if err != nil {
		log.Panic(err)
	}

	/*start web server*/
	go http.Serve(listener, s.gin.Handler())

	/*start  rtmp server*/
	s.rtmp.Run(ctx)

}

func (s *Group) HandleChannels(c *gin.Context) {

	res := channelResponse{}
	for _, ch := range s.rtmp.Channels {
		res.Channel = append(res.Channel, ch.Name())
	}
	c.JSON(http.StatusOK, res)

}

func (s *Group) HandleCreatePeerConn(c *gin.Context) {

	req := CreatePeerConnectRequest{}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		klog.Error(err)
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "Decoder Sdp error",
		})
	}

	app := req.Url
	ch, ok := s.rtmp.Channels[app]
	if !ok {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "app Not Found",
		})
	}

	/*Determine whether the channel already has an RTC player binding,
	if there is no one, if there is a single RTC playback to the playback group*/
	playerGroup, err := ch.LoadPlayer(rtc.RtcPlayerGroup)
	if err != nil {
		playerGroup = rtc.NewPlayer()
		ch.AddPlayer(rtc.RtcPlayerGroup, playerGroup)
	}

	player, err := rtc.NewRtcPlayer(req.Offer)
	if err != nil {
		klog.Error(err)
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "create rtc player error",
		})
	}

	playerGroup.(*rtc.Player).AddRtc(player)

	/*Return to the browser RTC properties*/
	response, err := json.Marshal(player.PeerConnection.CurrentLocalDescription())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "response Marshal",
		})
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	if _, err := c.Writer.Write(response); err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "response SessionDescription",
		})
	}

}

package sfu

import "github.com/pion/webrtc/v3"

type Room struct {
	SID             string
	PeerConnections []PeerConnectionState
	TrackLocals     map[string]*webrtc.TrackLocalStaticRTP
}

type JoinRoom struct {
	SID   string `json:"sid"`
	Token string `json:"token"`
}

type Conference struct {
	Rooms map[string]*Room
}

func (c *Conference) JoinRoom(sid string, peer PeerConnectionState) {
	room, ok := c.Rooms[sid]
	if ok {
		room.PeerConnections = append(room.PeerConnections, peer)
		return
	}
	c.Rooms[sid] = &Room{
		SID: sid,
		PeerConnections: []PeerConnectionState{
			peer,
		},
		TrackLocals: make(map[string]*webrtc.TrackLocalStaticRTP),
	}
}

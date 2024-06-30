package sfu

import (
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

var (
	TrackLocals = map[string]map[string]*webrtc.TrackLocalStaticRTP{
		"room": make(map[string]*webrtc.TrackLocalStaticRTP),
	}
)

// Add to list of tracks and fire renegotation for all PeerConnections
func AddTrack(t *webrtc.TrackRemote, room string) *webrtc.TrackLocalStaticRTP {
	ListLock.Lock()
	defer func() {
		ListLock.Unlock()
		SignalPeerConnections(room)
	}()

	// Create a new TrackLocal with the same codec as our incoming
	trackLocal, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, t.ID(), t.StreamID())
	if err != nil {
		panic(err)
	}

	_, ok := TrackLocals[room]

	if !ok {
		TrackLocals[room] = make(map[string]*webrtc.TrackLocalStaticRTP)
	}

	TrackLocals[room][t.ID()] = trackLocal
	return trackLocal
}

// Remove from list of tracks and fire renegotation for all PeerConnections
func RemoveTrack(t *webrtc.TrackLocalStaticRTP, room string) {
	ListLock.Lock()
	defer func() {
		ListLock.Unlock()
		SignalPeerConnections(room)
	}()

	delete(TrackLocals[room], t.ID())
}

// dispatchKeyFrame sends a keyframe to all PeerConnections, used everytime a new user joins the call
func DispatchKeyFrame(room string) {
	ListLock.Lock()
	defer ListLock.Unlock()

	for i := range PeerConnections[room] {
		for _, receiver := range PeerConnections[room][i].PeerConnection.GetReceivers() {
			if receiver.Track() == nil {
				continue
			}

			_ = PeerConnections[room][i].PeerConnection.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					MediaSSRC: uint32(receiver.Track().SSRC()),
				},
			})
		}
	}
}

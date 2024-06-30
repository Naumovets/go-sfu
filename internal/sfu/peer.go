package sfu

import (
	"encoding/json"
	"time"

	"github.com/pion/webrtc/v3"
)

type PeerConnectionState struct {
	PeerConnection *webrtc.PeerConnection
	Websocket      *ThreadSafeWriter
}

var (
	PeerConnections = make(map[string][]PeerConnectionState)
)

func SignalPeerConnections(room string) {
	ListLock.Lock()
	defer func() {
		ListLock.Unlock()
		DispatchKeyFrame(room)
	}()

	attemptSync := func() (tryAgain bool) {
		for i := range PeerConnections[room] {
			if PeerConnections[room][i].PeerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed {
				PeerConnections[room] = append(PeerConnections[room][:i], PeerConnections[room][i+1:]...)
				return true // We modified the slice, start from the beginning
			}

			// map of sender we already are seanding, so we don't double send
			existingSenders := map[string]bool{}

			for _, sender := range PeerConnections[room][i].PeerConnection.GetSenders() {
				if sender.Track() == nil {
					continue
				}

				existingSenders[sender.Track().ID()] = true

				// If we have a RTPSender that doesn't map to a existing track remove and signal
				if _, ok := TrackLocals[room][sender.Track().ID()]; !ok {
					if err := PeerConnections[room][i].PeerConnection.RemoveTrack(sender); err != nil {
						return true
					}
				}
			}

			// Don't receive videos we are sending, make sure we don't have loopback
			for _, receiver := range PeerConnections[room][i].PeerConnection.GetReceivers() {
				if receiver.Track() == nil {
					continue
				}

				existingSenders[receiver.Track().ID()] = true
			}

			// Add all track we aren't sending yet to the PeerConnection
			for trackID := range TrackLocals[room] {
				if _, ok := existingSenders[trackID]; !ok {
					if _, err := PeerConnections[room][i].PeerConnection.AddTrack(TrackLocals[room][trackID]); err != nil {
						return true
					}
				}
			}

			offer, err := PeerConnections[room][i].PeerConnection.CreateOffer(nil)
			if err != nil {
				return true
			}

			if err = PeerConnections[room][i].PeerConnection.SetLocalDescription(offer); err != nil {
				return true
			}

			offerString, err := json.Marshal(offer)
			if err != nil {
				return true
			}

			if err = PeerConnections[room][i].Websocket.WriteJSON(&WebsocketMessage{
				Event: "offer",
				Data:  string(offerString),
			}); err != nil {
				return true
			}
		}

		return
	}

	for syncAttempt := 0; ; syncAttempt++ {
		if syncAttempt == 25 {
			// Release the lock and attempt a sync in 3 seconds. We might be blocking a RemoveTrack or AddTrack
			go func() {
				time.Sleep(time.Second * 3)
				SignalPeerConnections(room)
			}()
			return
		}

		if !attemptSync() {
			break
		}
	}
}

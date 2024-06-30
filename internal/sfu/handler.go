package sfu

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/pion/webrtc/v3"
)

// Handle incoming websockets
func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP request to Websocket
	unsafeConn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	room := ""

	c := &ThreadSafeWriter{unsafeConn, sync.Mutex{}}

	// When this frame returns close the Websocket
	defer c.Close() //nolint

	// Create new PeerConnection
	PeerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Print(err)
		return
	}

	// When this frame returns close the PeerConnection
	defer PeerConnection.Close() //nolint

	message := &WebsocketMessage{}
	for {
		_, raw, err := c.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		} else if err := json.Unmarshal(raw, &message); err != nil {
			log.Println(err)
			return
		}

		switch message.Event {

		case "join":

			join := JoinRoom{}
			if err := json.Unmarshal([]byte(message.Data), &join); err != nil {
				log.Println(err)
				return
			}

			// user, ok := Auth(join.Token)

			// if !ok {
			// 	log.Println("User no authorized")
			// 	return
			// }
			// if ok {
			// 	log.Printf("user: %s %s\n", user.Name, user.Lastname)
			// }

			room = join.SID

			// Accept one audio and one video track incoming
			for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
				if _, err := PeerConnection.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
					Direction: webrtc.RTPTransceiverDirectionRecvonly,
				}); err != nil {
					log.Print(err)
					return
				}
			}

			// Add our new PeerConnection to global list
			ListLock.Lock()
			PeerConnections[room] = append(PeerConnections[room], PeerConnectionState{PeerConnection, c})
			ListLock.Unlock()

			// Trickle ICE. Emit server candidate to client
			PeerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
				if i == nil {
					return
				}

				candidateString, err := json.Marshal(i.ToJSON())
				if err != nil {
					log.Println(err)
					return
				}

				if writeErr := c.WriteJSON(&WebsocketMessage{
					Event: "candidate",
					Data:  string(candidateString),
				}); writeErr != nil {
					log.Println(writeErr)
				}
			})

			// If PeerConnection is closed remove it from global list
			PeerConnection.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
				switch p {
				case webrtc.PeerConnectionStateFailed:
					if err := PeerConnection.Close(); err != nil {
						log.Print(err)
					}
				case webrtc.PeerConnectionStateClosed:
					SignalPeerConnections(room)
				default:
				}
			})

			PeerConnection.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
				// Create a track to fan out our incoming video to all peers
				trackLocal := AddTrack(t, room)
				defer RemoveTrack(trackLocal, room)

				buf := make([]byte, 1500)
				for {
					i, _, err := t.Read(buf)
					if err != nil {
						return
					}

					if _, err = trackLocal.Write(buf[:i]); err != nil {
						return
					}
				}
			})

			// Signal for the new PeerConnection
			SignalPeerConnections(room)

		case "candidate":
			candidate := webrtc.ICECandidateInit{}
			if err := json.Unmarshal([]byte(message.Data), &candidate); err != nil {
				log.Println(err)
				return
			}

			if err := PeerConnection.AddICECandidate(candidate); err != nil {
				log.Println(err)
				return
			}
		case "answer":
			answer := webrtc.SessionDescription{}
			if err := json.Unmarshal([]byte(message.Data), &answer); err != nil {
				log.Println(err)
				return
			}

			if err := PeerConnection.SetRemoteDescription(answer); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

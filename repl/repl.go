package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/pion/webrtc/v3"
)

const DefaultRendezvous = "https://rendezvous.appinventor.mit.edu/rendezvous/"

type Repl struct {
	code         string
	sha1Digest   string
	rendezvous   string
	pollTimes    int
	onConnect    func(c *webrtc.DataChannel)
	onDisconnect func(graceful bool)
	onMessage    func(message webrtc.DataChannelMessage)

	peer    *webrtc.PeerConnection
	channel *webrtc.DataChannel
}

func NewRepl(
	code string,
	rendezvous string,
	pollTimes int,
	onConnect func(c *webrtc.DataChannel),
	onDisconnect func(canRecover bool),
	onMessage func(message webrtc.DataChannelMessage)) *Repl {
	// SHA1 Digest on the code
	sha1Hasher := sha1.New()
	sha1Hasher.Write([]byte(code))
	return &Repl{
		code:       code,
		sha1Digest: hex.EncodeToString(sha1Hasher.Sum(nil)),
		rendezvous: rendezvous,
		pollTimes:  pollTimes,

		onConnect:    onConnect,
		onDisconnect: onDisconnect,
		onMessage:    onMessage,
	}
}

func (r *Repl) Connect() error {
	pollUrl := r.rendezvous + r.sha1Digest
	for i := 0; i < r.pollTimes; i++ {
		// Await Companion at a rendezvous
		resp, err := http.Get(pollUrl)
		if err != nil {
			return err
		}
		if resp.StatusCode == 200 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			if len(body) != 0 {
				fmt.Println("Companion discovered, starting WebRTC...")
				return r.createComm(body)
			} else {
				fmt.Printf("(%d/%d) Waiting for companion\n", i, r.pollTimes)
			}
		} else {
			panic("First rendezvous poll failed: " + resp.Status)
		}
		resp.Body.Close()
		time.Sleep(time.Second)
	}
	r.peer.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		switch s {
		case webrtc.PeerConnectionStateFailed:
			fmt.Println("Connection Failed! Closing peer connection.")
			r.peer.Close()
		case webrtc.PeerConnectionStateDisconnected:
			// The connection may recover
			fmt.Println("Connection Disconnected. Waiting to see if it recovers...")
			r.onDisconnect(false)
		case webrtc.PeerConnectionStateClosed:
			// The connection was closed gracefully (on demand)
			r.onDisconnect(true)
		default:
			fmt.Println("Unknown connection state: " + s.String())
		}
	})
	// TODO:
	//  Listen for changes in network status
	return nil
}

type CommConfig struct {
	IceServers  []IceServer `json:"iceServers"`
	Rendezvous2 string      `json:"rendezvous2"`
}

type IceServer struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *Repl) createComm(body []byte) error {
	// deserialize whatever companion sent us
	var commConfig CommConfig
	if err := json.Unmarshal(body, &commConfig); err != nil {
		return err
	}
	// map IceServer struct to rtc ICEServer
	var rtcServers []webrtc.ICEServer
	for _, cfgServer := range commConfig.IceServers {
		rtcServer := webrtc.ICEServer{
			URLs:       []string{cfgServer.Server},
			Username:   cfgServer.Username,
			Credential: cfgServer.Password,
		}
		rtcServers = append(rtcServers, rtcServer)
	}
	// Init reconnection
	rtcConfig := webrtc.Configuration{ICEServers: rtcServers}
	peer, err := webrtc.NewPeerConnection(rtcConfig)
	if err != nil {
		return err
	}
	r.peer = peer
	// Retrieve channels and setup listeners
	ordered := true
	channel, err := r.peer.CreateDataChannel("data", &webrtc.DataChannelInit{
		Ordered: &ordered,
	})
	if err != nil {
		return err
	}
	r.channel = channel
	channel.OnOpen(func() {
		r.onConnect(channel)
	})
	channel.OnMessage(r.onMessage)

	// Send ICE Candidates to Companion one by one
	peer.OnICECandidate(func(c *webrtc.ICECandidate) {
		// Indicates candidates gathering completed
		if c == nil {
			return
		}
		content := map[string]interface{}{
			"key":       r.sha1Digest + "-s",
			"webrtc":    true,
			"nonce":     rand.Intn(10000) + 1,
			"candidate": c.ToJSON(),
		}
		jsonBytes, err := json.Marshal(content)
		if err != nil {
			panic(err)
		}
		resp, err := http.Post(commConfig.Rendezvous2, "application/json", bytes.NewBuffer(jsonBytes))
		if err != nil {
			panic(err)
		}
		resp.Body.Close()
		fmt.Printf("Sent ICE Candidate: %s\n", string(jsonBytes))
	})

	// Create an offer
	offer, err := peer.CreateOffer(nil)
	if err != nil {
		return err
	}
	err = peer.SetLocalDescription(offer)
	if err != nil {
		return err
	}
	offerContent := map[string]interface{}{
		"key":       r.code + "-s",
		"webrtc":    true,
		"offer":     offer,
		"nonce":     rand.Intn(10000) + 1,
		"candidate": nil,
	}
	jsonBytes, err := json.Marshal(offerContent)
	if err != nil {
		return err
	}
	// Post the offer
	resp, err := http.Post(commConfig.Rendezvous2, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}
	if resp.StatusCode == 200 {
		fmt.Println("WebRTC Offer Created. Waiting for an answer.")
	} else {
		fmt.Println("Failed to post WebRTC Offer, status: " + resp.Status)
	}
	resp.Body.Close()
	return r.receiveOfferResponse(commConfig.Rendezvous2 + r.code + "-r")
}

func (r *Repl) receiveOfferResponse(responseURL string) error {
	var pendingCandidates []webrtc.ICECandidateInit
	answerSet := false

	for i := 0; i < r.pollTimes; i++ {
		if r.peer.ConnectionState() == webrtc.PeerConnectionStateConnected {
			fmt.Println("peer connected!")
			break
		}
		print("(" + strconv.Itoa(i) + "/" + strconv.Itoa(r.pollTimes) + ") Waiting for SDP answer")

		resp, err := http.Get(responseURL)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			resp.Body.Close()
			continue
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			return err
		}
		resp.Body.Close()
		if len(body) == 0 {
			time.Sleep(time.Second)
			continue
		}
		var hunks []map[string]interface{}
		if err := json.Unmarshal(body, &hunks); err != nil {
			return err
		}
		for _, hunk := range hunks {
			// Check for an SDP answer
			if answerData, ok := hunk["offer"].(map[string]interface{}); ok && !answerSet {
				// Read offer response
				answerBytes, _ := json.Marshal(answerData)
				var answer webrtc.SessionDescription
				if err := json.Unmarshal(answerBytes, &answer); err != nil {
					return err
				}
				// Set remote descriptor
				if err := r.peer.SetRemoteDescription(answer); err != nil {
					return err
				}
				answerSet = true
				// Add any buffered (early received) candidates
				for _, candid := range pendingCandidates {
					if err := r.peer.AddICECandidate(candid); err != nil {
						return err
					}
				}
				pendingCandidates = nil
			}
			// Check for an ICE Candidate
			if candidData, ok := hunk["candidate"].(map[string]interface{}); ok {
				candidBytes, _ := json.Marshal(candidData)
				var candid webrtc.ICECandidateInit
				if err := json.Unmarshal(candidBytes, &candid); err != nil {
					return err
				}
				if answerSet {
					if err := r.peer.AddICECandidate(candid); err != nil {
						return err
					}
				} else {
					// SDP Answer not yet received, buffer for now
					pendingCandidates = append(pendingCandidates, candid)
				}
			}
		}
		time.Sleep(time.Second)
	}
	return nil
}

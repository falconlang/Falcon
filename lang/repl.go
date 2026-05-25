//go:build !js && !wasm

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
	sha1Hasher := sha1.New()
	sha1Hasher.Write([]byte(code))
	return &Repl{
		code:         code,
		sha1Digest:   hex.EncodeToString(sha1Hasher.Sum(nil)),
		rendezvous:   rendezvous,
		pollTimes:    pollTimes,
		onConnect:    onConnect,
		onDisconnect: onDisconnect,
		onMessage:    onMessage,
	}
}

func (r *Repl) Connect() error {
	pollUrl := r.rendezvous + r.sha1Digest
	for i := 0; i < r.pollTimes; i++ {
		resp, err := http.Get(pollUrl)
		if err != nil {
			return err
		}
		if resp.StatusCode == 200 {
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return err
			}
			if len(body) != 0 {
				fmt.Println("Companion discovered, starting WebRTC...")
				return r.createComm(body)
			}
			fmt.Printf("(%d/%d) Waiting for companion\n", i+1, r.pollTimes)
		} else {
			resp.Body.Close()
			return fmt.Errorf("rendezvous poll failed: %s", resp.Status)
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("companion not found after %d attempts", r.pollTimes)
}

func (r *Repl) SendText(text string) error {
	return r.channel.SendText(text)
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
	var commConfig CommConfig
	if err := json.Unmarshal(body, &commConfig); err != nil {
		return err
	}

	var rtcServers []webrtc.ICEServer
	for _, cfgServer := range commConfig.IceServers {
		rtcServers = append(rtcServers, webrtc.ICEServer{
			URLs:       []string{cfgServer.Server},
			Username:   cfgServer.Username,
			Credential: cfgServer.Password,
		})
	}

	peer, err := webrtc.NewPeerConnection(webrtc.Configuration{ICEServers: rtcServers})
	if err != nil {
		return err
	}
	r.peer = peer

	peer.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		switch s {
		case webrtc.PeerConnectionStateFailed:
			fmt.Println("Connection failed, closing.")
			r.peer.Close()
		case webrtc.PeerConnectionStateDisconnected:
			fmt.Println("Connection disconnected.")
			r.onDisconnect(false)
		case webrtc.PeerConnectionStateClosed:
			r.onDisconnect(true)
		}
	})

	ordered := true
	channel, err := peer.CreateDataChannel("data", &webrtc.DataChannelInit{Ordered: &ordered})
	if err != nil {
		return err
	}
	r.channel = channel
	channel.OnOpen(func() { r.onConnect(channel) })
	channel.OnMessage(r.onMessage)

	peer.OnICECandidate(func(c *webrtc.ICECandidate) {
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
			fmt.Println("ICE marshal error:", err)
			return
		}
		resp, err := http.Post(commConfig.Rendezvous2, "application/json", bytes.NewBuffer(jsonBytes))
		if err != nil {
			fmt.Println("ICE post error:", err)
			return
		}
		resp.Body.Close()
	})

	offer, err := peer.CreateOffer(nil)
	if err != nil {
		return err
	}
	if err = peer.SetLocalDescription(offer); err != nil {
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
	resp, err := http.Post(commConfig.Rendezvous2, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}
	if resp.StatusCode == 200 {
		fmt.Println("WebRTC offer sent. Waiting for answer...")
	} else {
		fmt.Println("Offer post failed:", resp.Status)
	}
	resp.Body.Close()

	return r.receiveOfferResponse(commConfig.Rendezvous2 + r.code + "-r")
}

func (r *Repl) receiveOfferResponse(responseURL string) error {
	var pendingCandidates []webrtc.ICECandidateInit
	answerSet := false

	for i := 0; i < r.pollTimes; i++ {
		if r.peer.ConnectionState() == webrtc.PeerConnectionStateConnected {
			fmt.Println("Peer connected!")
			break
		}
		fmt.Print("(" + strconv.Itoa(i+1) + "/" + strconv.Itoa(r.pollTimes) + ") Waiting for SDP answer\n")

		resp, err := http.Get(responseURL)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			resp.Body.Close()
			time.Sleep(time.Second)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}
		if len(body) == 0 {
			time.Sleep(time.Second)
			continue
		}

		var hunks []map[string]interface{}
		if err := json.Unmarshal(body, &hunks); err != nil {
			return err
		}

		for _, hunk := range hunks {
			if answerData, ok := hunk["offer"].(map[string]interface{}); ok && !answerSet {
				answerBytes, _ := json.Marshal(answerData)
				var answer webrtc.SessionDescription
				if err := json.Unmarshal(answerBytes, &answer); err != nil {
					return err
				}
				if err := r.peer.SetRemoteDescription(answer); err != nil {
					return err
				}
				answerSet = true
				for _, candid := range pendingCandidates {
					if err := r.peer.AddICECandidate(candid); err != nil {
						return err
					}
				}
				pendingCandidates = nil
			}
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
					pendingCandidates = append(pendingCandidates, candid)
				}
			}
		}
		time.Sleep(time.Second)
	}
	return nil
}

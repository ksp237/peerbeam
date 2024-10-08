package receiver

import (
	"github.com/6b70/peerbeam/conn"
	"github.com/6b70/peerbeam/proto/compiled/controlpb"
	"github.com/6b70/peerbeam/utils"
)

type Receiver struct {
	*conn.Session
}

func New() *Receiver {
	return &Receiver{
		Session: conn.New(),
	}
}

func (r *Receiver) SetupReceiverConn() error {
	err := r.SetupPeerConn()
	if err != nil {
		return err
	}
	r.addChHandler()
	return nil
}

func (r *Receiver) CreateAnswer(offer string) (string, error) {
	offerSDP, err := utils.DecodeSDP(offer)
	if err != nil {
		return "", err
	}
	err = r.AddRemote(offerSDP)
	if err != nil {
		return "", err
	}

	answer, err := r.createSDP()
	if err != nil {
		return "", err
	}

	return answer, nil
}

func (r *Receiver) Receive(fileMDList *controlpb.FileMetadataList, destPath string) error {
	err := r.receiveFiles(fileMDList, destPath)
	if err != nil {
		return err
	}
	// Doesn't always flush but this is handled in the sender
	_ = r.DataCh.Close()
	return nil
}

func (r *Receiver) createSDP() (string, error) {
	initAnswer, err := r.Conn.CreateAnswer(nil)
	if err != nil {
		return "", err
	}
	err = r.Conn.SetLocalDescription(initAnswer)
	if err != nil {
		return "", err
	}

	r.CandidateCond.L.Lock()
	r.CandidateCond.Wait()
	r.CandidateCond.L.Unlock()

	answer := r.Conn.LocalDescription()

	encodedSDP, err := utils.EncodeSDP(answer)
	if err != nil {
		return "", err
	}

	return encodedSDP, nil
}

package compat

import (
	"crypto/sha1"
	"io"

	"github.com/twstrike/otr3"
)

// QueryMessage can be sent to a peer to start an OTR conversation.
var QueryMessage = "?OTRv2?"

// ErrorPrefix can be used to make an OTR error by appending an error message
// to it.
var ErrorPrefix = "?OTR Error:"

// SecurityChange describes a change in the security state of a Conversation.
type SecurityChange int

const (
	NoChange SecurityChange = iota
	// NewKeys indicates that a key exchange has completed. This occurs
	// when a conversation first becomes encrypted, and when the keys are
	// renegotiated within an encrypted conversation.
	NewKeys
	// SMPSecretNeeded indicates that the peer has started an
	// authentication and that we need to supply a secret. Call SMPQuestion
	// to get the optional, human readable challenge and then Authenticate
	// to supply the matching secret.
	SMPSecretNeeded
	// SMPComplete indicates that an authentication completed. The identity
	// of the peer has now been confirmed.
	SMPComplete
	// SMPFailed indicates that an authentication failed.
	SMPFailed
	// ConversationEnded indicates that the peer ended the secure
	// conversation.
	ConversationEnded
)

type Conversation struct {
	otr3.Conversation
	TheirPublicKey PublicKey
	PrivateKey     *PrivateKey
	SSID           [8]byte
	FragmentSize   int
}

func (c *Conversation) compatInit() {
	c.Conversation.Policies.AllowV2()
	c.OurKey = &c.PrivateKey.PrivateKey
	c.TheirKey = &c.TheirPublicKey.PublicKey
}

func (c *Conversation) Receive(in []byte) (out []byte, encrypted bool, change SecurityChange, toSend [][]byte, err error) {
	c.compatInit()
	encrypted = c.IsEncrypted()
	var ret []otr3.ValidMessage
	out, ret, err = c.Conversation.Receive(in)

	if ret != nil {
		toSend = otr3.Bytes(ret)
	}

	return
}

func (c *Conversation) Send(in []byte) (toSend [][]byte, err error) {
	c.compatInit()

	var ret []otr3.ValidMessage
	ret, err = c.Conversation.Send(in)

	if ret != nil {
		toSend = otr3.Bytes(ret)
	}

	return
}

func (c *Conversation) End() (toSend [][]byte) {
	c.compatInit()

	var ret []otr3.ValidMessage
	ret, _ = c.Conversation.End()

	if ret != nil {
		toSend = otr3.Bytes(ret)
	}

	return
}

func (c *Conversation) Authenticate(question string, mutualSecret []byte) (toSend [][]byte, err error) {
	c.compatInit()
	return [][]byte{}, nil
}

func (c *Conversation) SMPQuestion() string {
	c.compatInit()
	question, _ := c.Conversation.SMPQuestion()
	return question
}

type PublicKey struct {
	otr3.PublicKey
}

type PrivateKey struct {
	otr3.PrivateKey
}

func (priv *PrivateKey) Generate(rand io.Reader) {
	if err := priv.PrivateKey.Generate(rand); err != nil {
		panic(err.Error())
	}
}

func (priv *PrivateKey) Serialize(in []byte) []byte {
	return append(in, priv.PrivateKey.Serialize()...)
}

func (priv *PrivateKey) Fingerprint() []byte {
	return priv.PublicKey.Fingerprint(sha1.New())
}

func (pub *PublicKey) Fingerprint() []byte {
	return pub.PublicKey.Fingerprint(sha1.New())
}

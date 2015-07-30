package otr3

import (
	"crypto/rand"
	"testing"
)

func Test_conversation_SMPStateMachineStartsAtSmpExpect1(t *testing.T) {
	c := newConversation(otrV3{}, fixtureRand())
	assertEquals(t, c.smp.state, smpStateExpect1{})
}

func Test_receive_AbortsSMPStateMachineIfDoesNotHaveASecureChannel(t *testing.T) {
	states := []msgState{
		plainText, finished,
	}

	c := bobContextAfterAKE()

	smpMsg := fixtureMessage1()
	m := c.genDataMsg(nil, smpMsg.tlv()).serialize(c)
	smpAbortMsg := smpMessageAbort{}.tlv().serialize()

	for _, s := range states {
		c.msgState = s

		_, toSend, err := c.Receive(m)
		assertEquals(t, err, errEncryptedMessageWithNoSecureChannel)
		assertEquals(t, c.smp.state, smpStateExpect1{})
		assertDeepEquals(t, toSend, smpAbortMsg)
	}
}

func Test_AKEHappyPath(t *testing.T) {
	alice := newConversation(otrV3{}, rand.Reader)
	bob := newConversation(otrV3{}, rand.Reader)
	alice.policies.add(allowV2)
	bob.policies.add(allowV2)
	alice.policies.add(allowV3)
	bob.policies.add(allowV3)
	alice.ourKey = alicePrivateKey
	bob.ourKey = bobPrivateKey
	alice.theirKey = &bobPrivateKey.PublicKey
	bob.theirKey = &alicePrivateKey.PublicKey

	msg := []byte("?OTRv3?")
	var toSend []byte
	var err error
	//Alice send Bob queryMsg
	_, toSend, err = bob.Receive(msg)
	assertEquals(t, err, nil)
	assertEquals(t, bob.ake.state, authStateAwaitingDHKey{})

	//Bob send Alice DHCommit
	_, toSend, err = alice.Receive(toSend)
	assertEquals(t, alice.ake.state, authStateAwaitingRevealSig{})
	assertEquals(t, err, nil)

	//Alice send Bob DHKey
	_, toSend, err = bob.Receive(toSend)
	assertEquals(t, err, nil)
	assertDeepEquals(t, bob.ake.state, authStateAwaitingSig{revealSigMsg: toSend})

	//Bob send Alice RevealSig
	_, toSend, err = alice.Receive(toSend)
	assertEquals(t, err, nil)
	assertEquals(t, alice.ake.state, authStateNone{})

	//Alice send Bob Sig
	_, toSend, err = bob.Receive(toSend)
	assertEquals(t, err, nil)
	assertEquals(t, bob.ake.state, authStateNone{})

	// "When starting a private Conversation [...],
	// generate two DH key pairs for yourself, and set our_keyid = 2"
	assertEquals(t, alice.keys.ourKeyID, uint32(2))
	assertEquals(t, alice.keys.ourCurrentDHKeys.priv.BitLen() > 0, true)
	assertEquals(t, alice.keys.ourCurrentDHKeys.pub.BitLen() > 0, true)
	assertEquals(t, alice.keys.ourPreviousDHKeys.priv.BitLen() > 0, true)
	assertEquals(t, alice.keys.ourPreviousDHKeys.pub.BitLen() > 0, true)

	assertEquals(t, bob.keys.ourKeyID, uint32(2))
	assertEquals(t, bob.keys.ourCurrentDHKeys.priv.BitLen() > 0, true)
	assertEquals(t, bob.keys.ourCurrentDHKeys.pub.BitLen() > 0, true)
	assertEquals(t, bob.keys.ourPreviousDHKeys.priv.BitLen() > 0, true)
	assertEquals(t, bob.keys.ourPreviousDHKeys.pub.BitLen() > 0, true)
}

func Test_AKENotAllowV2(t *testing.T) {
	alice := newConversation(otrV3{}, rand.Reader)
	bob := newConversation(otrV3{}, rand.Reader)
	alice.policies.add(allowV3)
	bob.policies.add(allowV3)
	alice.ourKey = alicePrivateKey
	bob.ourKey = bobPrivateKey
	alice.theirKey = &bobPrivateKey.PublicKey
	bob.theirKey = &alicePrivateKey.PublicKey

	msg := []byte("?OTRv3?")
	var toSend []byte
	var nilB []byte
	var err error
	//Alice send Bob queryMsg
	_, toSend, err = bob.Receive(msg)
	assertEquals(t, err, nil)
	assertEquals(t, bob.ake.state, authStateAwaitingDHKey{})

	//Bob send Alice DHCommit
	_, toSend, err = alice.Receive(toSend)
	assertEquals(t, alice.ake.state, authStateAwaitingRevealSig{})
	assertEquals(t, err, nil)

	//Alice send Bob DHKey
	_, toSend, err = bob.Receive(toSend)
	assertEquals(t, err, nil)
	assertDeepEquals(t, bob.ake.state, authStateAwaitingSig{revealSigMsg: toSend})

	//Bob send Alice RevealSig
	_, toSend, err = alice.Receive(toSend)
	assertEquals(t, err, nil)
	assertEquals(t, alice.ake.state, authStateAwaitingRevealSig{})
	assertDeepEquals(t, toSend, nilB)
}

func Test_processDataMessageShouldExtractData(t *testing.T) {
	var toSend []byte
	var err error
	var nilB []byte

	alice := newConversation(nil, rand.Reader)
	alice.policies = policies(allowV2 | allowV3)
	alice.ourKey = alicePrivateKey

	bob := newConversation(nil, rand.Reader)
	bob.policies = policies(allowV2 | allowV3)
	bob.ourKey = bobPrivateKey

	msg := []byte("?OTRv3?")

	//Alice send Bob queryMsg
	_, toSend, err = bob.Receive(msg)
	assertEquals(t, err, nil)
	assertEquals(t, bob.ake.state, authStateAwaitingDHKey{})
	assertEquals(t, bob.version, otrV3{})

	//Bob send Alice DHCommit
	_, toSend, err = alice.Receive(toSend)
	assertEquals(t, alice.ake.state, authStateAwaitingRevealSig{})
	assertEquals(t, err, nil)

	//Alice send Bob DHKey
	_, toSend, err = bob.Receive(toSend)
	assertEquals(t, err, nil)
	assertDeepEquals(t, bob.ake.state, authStateAwaitingSig{revealSigMsg: toSend})

	//Bob send Alice RevealSig
	_, toSend, err = alice.Receive(toSend)
	assertEquals(t, err, nil)
	assertEquals(t, alice.ake.state, authStateNone{})

	//Alice send Bob Sig
	_, toSend, err = bob.Receive(toSend)
	assertEquals(t, err, nil)
	assertEquals(t, bob.ake.state, authStateNone{})

	// Alice sends a message to bob
	m := []byte("hello")
	datamsg := alice.genDataMsg(m).serialize(alice)
	plain, toSend, err := bob.processDataMessage(datamsg)

	assertDeepEquals(t, err, nil)
	assertDeepEquals(t, plain, m)
	assertDeepEquals(t, toSend, nilB)
}

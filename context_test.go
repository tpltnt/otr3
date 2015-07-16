package otr3

import "testing"

var (
	fixtureX  = bnFromHex("abcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcd")
	fixtureGx = bnFromHex("2cdacabb00e63d8949aa85f7e6a095b1ee81a60779e58f8938ff1a7ed1e651d954bd739162e699cc73b820728af53aae60a46d529620792ddf839c5d03d2d4e92137a535b27500e3b3d34d59d0cd460d1f386b5eb46a7404b15c1ef84840697d2d3d2405dcdda351014d24a8717f7b9c51f6c84de365fea634737ae18ba22253a8e15249d9beb2dded640c6c0d74e4f7e19161cf828ce3ffa9d425fb68c0fddcaa7cbe81a7a5c2c595cce69a255059d9e5c04b49fb15901c087e225da850ff27")
)

func Test_parseOTRQueryMessage(t *testing.T) {
	var exp = map[string][]int{
		"?OTR?":     []int{1},
		"?OTRv2?":   []int{2},
		"?OTRv23?":  []int{2, 3},
		"?OTR?v2":   []int{1, 2},
		"?OTRv248?": []int{2, 4, 8},
		"?OTR?v?":   []int{1},
		"?OTRv?":    []int{},
	}

	for queryMsg, versions := range exp {
		m := []byte(queryMsg)
		assertDeepEquals(t, parseOTRQueryMessage(m), versions)
	}
}

func Test_receiveSendsDHCommitMessageAfterReceivingAnOTRQueryMessage(t *testing.T) {
	msg := []byte("?OTRv3?")
	cxt := newContext(otrV3{}, fixtureRand())

	exp := []byte{
		0x00, 0x03, // protocol version
		0x02, //DH message type
	}

	toSend, err := cxt.receive(msg)

	assertEquals(t, err, nil)
	assertDeepEquals(t, toSend[:3], exp)
}

func Test_receiveVerifiesMessageProtocolVersion(t *testing.T) {
	// protocol version
	msg := []byte{0x00, 0x02}
	cxt := newContext(otrV3{}, fixtureRand())

	_, err := cxt.receive(msg)
	assertEquals(t, err, errWrongProtocolVersion)
}

func Test_receiveDHCommitMessageReturnsDHKeyForOTR3(t *testing.T) {
	exp := []byte{
		0x00, 0x03, // protocol version
		0x0A, //DH message type
	}

	dhCommitAKE := fixtureAKE()
	dhCommitMsg, _ := dhCommitAKE.dhCommitMessage()
	cxt := newContext(otrV3{}, fixtureRand())
	ake := AKE{}
	ake.otrVersion = otrV3{}

	dhKeyMsg, err := cxt.receive(dhCommitMsg)

	assertEquals(t, err, nil)
	assertDeepEquals(t, dhKeyMsg[:lenMsgHeader], exp)
}

func Test_receiveDHKeyMessageGeneratesDHRevealSigMessage(t *testing.T) {
	exp := []byte{
		0x00, 0x03, // protocol version
		msgTypeRevealSig, // type
	}

	cxt := newContext(otrV3{}, fixtureRand())
	cxt.x = fixtureX
	cxt.gx = fixtureGx
	cxt.privateKey = bobPrivateKey

	ake := AKE{}
	ake.otrVersion = otrV3{}

	dhKeyMsg, _ := ake.dhKeyMessage()

	dhRevealSigMsg, err := cxt.receiveDHKey(dhKeyMsg)
	assertEquals(t, err, nil)
	assertDeepEquals(t, dhRevealSigMsg[:lenMsgHeader], exp)
}

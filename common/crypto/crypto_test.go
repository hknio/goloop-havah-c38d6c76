package crypto

import (
	"encoding/hex"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/icon-project/goloop/common/codec"

	"github.com/stretchr/testify/assert"
)

var (
	testHash             = SHA3Sum256([]byte("icx_sendTransaction.fee.0x2386f26fc10000.from.hx57b8365292c115d3b72d948272cc4d788fa91f64.timestamp.1538976759263551.to.hx57b8365292c115d3b72d948272cc4d788fa91f64.value.0xde0b6b3a7640000"))
	testPrivateKey, _    = hex.DecodeString("ca158b1d3c81c492e7785a3bba6aa755e07c28d2711811e7014bcf911ea2643b")
	testPublicKey, _     = hex.DecodeString("0448250ebe88d77e0a12bcf530fe6a2cf1ac176945638d309b840d631940c93b78c2bd6d16f227a8877e3f1604cd75b9c5a8ab0cac95174a8a0a0f8ea9e4c10bca")
	testPublicKeyComp, _ = hex.DecodeString("0248250ebe88d77e0a12bcf530fe6a2cf1ac176945638d309b840d631940c93b78")
	testSignature, _     = hex.DecodeString("4011de30c04302a2352400df3d1459d6d8799580dceb259f45db1d99243a8d0c64f548b7776cb93e37579b830fc3efce41e12e0958cda9f8c5fcad682c61079500")
)

// TODO add performance test
func TestSignAndVerify(t *testing.T) {
	priv, pub := GenerateKeyPair()
	sig, err := NewSignature(testHash, priv)
	if err != nil {
		t.Errorf("error signing:%s", err)
		return
	}

	if !sig.Verify(testHash, pub) {
		t.Errorf("Verify failed")
	}

	hash := make([]byte, len(testHash))
	copy(hash, testHash)
	hash[0] ^= 0xff
	if sig.Verify(hash, pub) {
		t.Errorf("Verify always works!")
	}
}

func TestVerifySignature(t *testing.T) {
	sig, _ := ParseSignature(testSignature)
	pub, _ := ParsePublicKey(testPublicKey)
	// pub, _ := ParsePublicKey(testPublicKeyComp)
	if !sig.Verify(testHash, pub) {
		t.Errorf("Verify failed")
	}
}

func TestRecoverPublicKey(t *testing.T) {
	priv, pub := GenerateKeyPair()
	sig, err := NewSignature(testHash, priv)

	pub1, err := sig.RecoverPublicKey(testHash)
	if err != nil {
		t.Errorf("error recover public key:%s", err)
		return
	}

	if !pub.Equal(pub1) {
		t.Errorf("recovered public key is not same")
	}

	sig.bytes[0] = sig.bytes[0] ^ 0x0f
	pub2, err := sig.RecoverPublicKey(testHash)
	if err == nil && pub.Equal(pub2) {
		t.Errorf("Public key recovery always works!")
	}
}

func TestPrintSignature(t *testing.T) {
	sig, _ := ParseSignature(testSignature)
	str := "0x" + hex.EncodeToString(testSignature)
	if strings.Compare(sig.String(), str) != 0 {
		t.Errorf("fail to print signature")
	}

	sig, _ = ParseSignature(testSignature[:64])
	str = "0x" + hex.EncodeToString(testSignature[:64]) + "[no V]"
	if strings.Compare(sig.String(), str) != 0 {
		t.Errorf("fail to print signaure(no V)")
	}

	sig, _ = ParseSignature([]byte("invalid"))
	str = "[empty]"
	if strings.Compare(sig.String(), str) != 0 {
		t.Errorf("fail to print signaure(no V)")
	}
}

func TestSignature_RLPEncodeSelf(t *testing.T) {
	priv, pub := GenerateKeyPair()
	sig, err := NewSignature(testHash, priv)
	assert.NoError(t, err)

	bs := codec.MustMarshalToBytes(&sig)
	var sig2 Signature
	codec.MustUnmarshalFromBytes(bs, &sig2)
	rpub, err := sig.RecoverPublicKey(testHash)
	assert.NoError(t, err)
	rpub2, err := sig2.RecoverPublicKey(testHash)
	assert.NoError(t, err)

	assert.EqualValues(t, pub.SerializeCompressed(), rpub.SerializeCompressed())
	assert.EqualValues(t, rpub.SerializeCompressed(), rpub2.SerializeCompressed())
}

func TestRace(t *testing.T) {
	const SubRoutineCount = 4
	const RepeatCount = 5

	var lock sync.Mutex
	cond := sync.NewCond(&lock)

	var readyWG sync.WaitGroup
	readyWG.Add(SubRoutineCount)
	wait := func() {
		lock.Lock()
		defer lock.Unlock()
		readyWG.Done()
		cond.Wait()
	}

	startAll := func() {
		lock.Lock()
		defer lock.Unlock()
		cond.Broadcast()
	}

	var finishWG sync.WaitGroup
	subRoutine := func(idx int) {
		wait()
		for i := 0; i < RepeatCount; i++ {
			priv, pub := GenerateKeyPair()
			sig, err := NewSignature(testHash, priv)

			pub1, err := sig.RecoverPublicKey(testHash)
			if err != nil {
				t.Errorf("error recover public key:%s", err)
				return
			}

			if !pub.Equal(pub1) {
				t.Errorf("recovered public key is not same")
			}
			r := sig.Verify(testHash, pub)
			assert.True(t, r)
			delay := time.Millisecond * time.Duration(rand.Intn(10))
			time.Sleep(delay)
		}
		finishWG.Done()
	}

	// start subroutines
	finishWG.Add(SubRoutineCount)
	for i := 0; i < SubRoutineCount; i++ {
		go subRoutine(i)
	}

	// wait until the subroutines reach wait()
	readyWG.Wait()
	time.Sleep(10 * time.Millisecond)

	// start subroutines
	startAll()

	// wait for DONE
	finishWG.Wait()
}

package blsbft

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/incognitokey"

	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/consensus/signatureschemes/bridgesig"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
)

func TestMiningKey_GetKeyTuble(t *testing.T) {
	lenOutput := 10
	for j := 0; j < common.MaxShardNumber; j++ {
		privKeyLs := make([]string, 0)
		paymentAddLs := make([]string, 0)
		miningSeedLs := make([]string, 0)
		publicKeyLs := make([]string, 0)
		committeeKeyLs := make([]string, 0)
		for i := 0; i < 10000; i++ {
			seed := privacy.RandomScalar().ToBytesS()
			masterKey, _ := wallet.NewMasterKey(seed)
			child, _ := masterKey.NewChildKey(uint32(i))
			privKeyB58 := child.Base58CheckSerialize(wallet.PriKeyType)
			paymentAddressB58 := child.Base58CheckSerialize(wallet.PaymentAddressType)
			shardID := common.GetShardIDFromLastByte(child.KeySet.PaymentAddress.Pk[len(child.KeySet.PaymentAddress.Pk)-1])
			miningSeed := base58.Base58Check{}.Encode(common.HashB(common.HashB(child.KeySet.PrivateKey)), common.ZeroByte)
			publicKey := base58.Base58Check{}.Encode(child.KeySet.PaymentAddress.Pk, common.ZeroByte)
			committeeKey, _ := incognitokey.NewCommitteeKeyFromSeed(common.HashB(common.HashB(child.KeySet.PrivateKey)), child.KeySet.PaymentAddress.Pk)

			//viewingKeyB58 := child.Base58CheckSerialize(wallet.ReadonlyKeyType)
			//publicKeyB58 := child.KeySet.GetPublicKeyInBase58CheckEncode()

			//fmt.Println("privKeyB58: ", privKeyB58)
			//fmt.Println("publicKeyB58: ", publicKeyB58)
			//fmt.Println("paymentAddressB58: ", paymentAddressB58)
			//fmt.Println("viewingKeyB58: ", viewingKeyB58)

			//blsBft := BLSBFT{}
			//privateSeed, _ := blsBft.LoadUserKeyFromIncPrivateKey(privKeyB58)

			//fmt.Println("privateSeed: ", privateSeed)
			//fmt.Println()
			if int(shardID) == j {

				privKeyLs = append(privKeyLs, strconv.Quote(privKeyB58))
				paymentAddLs = append(paymentAddLs, strconv.Quote(paymentAddressB58))
				miningSeedLs = append(miningSeedLs, strconv.Quote(miningSeed))
				publicKeyLs = append(publicKeyLs, strconv.Quote(publicKey))
				temp, _ := committeeKey.ToBase58()
				committeeKeyLs = append(committeeKeyLs, strconv.Quote(temp))
				if len(privKeyLs) >= lenOutput {
					break
				}
			}
		}
		fmt.Println("privKeyLs"+strconv.Itoa(j), " = [", strings.Join(privKeyLs, ", "), "]")
		fmt.Println("paymentAddLs"+strconv.Itoa(j), " = [", strings.Join(paymentAddLs, ", "), "]")
		fmt.Println("miningSeedLs"+strconv.Itoa(j), " = [", strings.Join(miningSeedLs, ", "), "]")
		fmt.Println("publicKeyLs"+strconv.Itoa(j), " = [", strings.Join(publicKeyLs, ", "), "]")
		fmt.Println("committeeKeyLs"+strconv.Itoa(j), " = [", strings.Join(committeeKeyLs, ", "), "]")
	}
}

func newMiningKey(privateSeed string) (*MiningKey, error) {
	var miningKey MiningKey
	privateSeedBytes, _, err := base58.Base58Check{}.Decode(privateSeed)
	if err != nil {
		return nil, NewConsensusError(LoadKeyError, err)
	}

	blsPriKey, blsPubKey := blsmultisig.KeyGen(privateSeedBytes)

	// privateKey := blsmultisig.B2I(privateKeyBytes)
	// publicKeyBytes := blsmultisig.PKBytes(blsmultisig.PKGen(privateKey))
	miningKey.PriKey = map[string][]byte{}
	miningKey.PubKey = map[string][]byte{}
	miningKey.PriKey[common.BlsConsensus] = blsmultisig.SKBytes(blsPriKey)
	miningKey.PubKey[common.BlsConsensus] = blsmultisig.PKBytes(blsPubKey)
	bridgePriKey, bridgePubKey := bridgesig.KeyGen(privateSeedBytes)
	miningKey.PriKey[common.BridgeConsensus] = bridgesig.SKBytes(&bridgePriKey)
	miningKey.PubKey[common.BridgeConsensus] = bridgesig.PKBytes(&bridgePubKey)
	return &miningKey, nil
}

func Test_newMiningKey(t *testing.T) {
	type args struct {
		privateSeed string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Get mining key from private seed",
			args: args{
				privateSeed: "1Md5Jd3syKLygiphTyXZGLQFswsbgPpVfchYfiVrHX86A6Zsyn",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if key, err := newMiningKey(tt.args.privateSeed); (err != nil) != tt.wantErr {
				t.Errorf("newMiningKey() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				fmt.Println("BLS Key:", base58.Base58Check{}.Encode(key.PubKey[common.BlsConsensus], common.Base58Version))
				fmt.Println("BRI Key:", base58.Base58Check{}.Encode(key.PubKey[common.BridgeConsensus], common.Base58Version))
			}
		})
	}
}

func TestGenerateKey(t *testing.T) {
	lenOutput := 22
	seed := []byte("safafsafafsafafsafafsafafsafafsafafsafafsafafsafafsafaf")
	for j := 0; j < common.MaxShardNumber; j++ {
		privKeyLs := make([]string, 0)
		paymentAddLs := make([]string, 0)
		miningSeedLs := make([]string, 0)
		publicKeyLs := make([]string, 0)
		committeeKeyLs := make([]string, 0)
		for i := 0; i < 1000; i++ {
			masterKey, _ := wallet.NewMasterKey(seed)
			child, _ := masterKey.NewChildKey(uint32(i))

			privKeyB58 := child.Base58CheckSerialize(wallet.PriKeyType)
			paymentAddressB58 := child.Base58CheckSerialize(wallet.PaymentAddressType)
			shardID := common.GetShardIDFromLastByte(child.KeySet.PaymentAddress.Pk[len(child.KeySet.PaymentAddress.Pk)-1])
			miningSeed := base58.Base58Check{}.Encode(common.HashB(common.HashB(child.KeySet.PrivateKey)), common.ZeroByte)
			publicKey := base58.Base58Check{}.Encode(child.KeySet.PaymentAddress.Pk, common.ZeroByte)
			committeeKey, _ := incognitokey.NewCommitteeKeyFromSeed(common.HashB(common.HashB(child.KeySet.PrivateKey)), child.KeySet.PaymentAddress.Pk)

			if int(shardID) == j {

				privKeyLs = append(privKeyLs, (privKeyB58))
				paymentAddLs = append(paymentAddLs, (paymentAddressB58))
				miningSeedLs = append(miningSeedLs, (miningSeed))
				publicKeyLs = append(publicKeyLs, (publicKey))
				temp, _ := committeeKey.ToBase58()
				committeeKeyLs = append(committeeKeyLs, (temp))
				if len(privKeyLs) >= lenOutput {
					break
				}
			}
		}

		fmt.Printf("\n\n\n ***** Shard %+v **** \n\n\n", j)
		for i := 0; i < len(privKeyLs); i++ {
			fmt.Println(i)
			fmt.Println("Private Key: " + privKeyLs[i])
			fmt.Println("Payment Address: " + paymentAddLs[i])
			fmt.Println("Public key: " + publicKeyLs[i])
			fmt.Println("Mining key: " + miningSeedLs[i])
			fmt.Println("Committee key set: " + committeeKeyLs[i])
			fmt.Println("------------------------------------------------------------")
		}
	}
}

func GenerateFullKeyFromPrivateKey(privateKey string) error {
	// generate inc key
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return err
	}
	err = keyWallet.KeySet.InitFromPrivateKeyByte(keyWallet.KeySet.PrivateKey)
	if err != nil {
		return err
	}

	// calculate private seed
	privateSeedBytes := common.HashB(common.HashB(keyWallet.KeySet.PrivateKey))
	committeePubKey, err := incognitokey.NewCommitteeKeyFromSeed(privateSeedBytes, keyWallet.KeySet.PaymentAddress.Pk)
	if err != nil {
		return err
	}
	committeeKeyStr, err := committeePubKey.ToBase58()
	if err != nil {
		return err
	}

	// print result
	privateKeyStr := keyWallet.Base58CheckSerialize(wallet.PriKeyType)
	paymentAddrStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)
	readOnlyKeyStr := keyWallet.Base58CheckSerialize(wallet.ReadonlyKeyType)

	privateSeedStr := base58.Base58Check{}.Encode(privateSeedBytes, common.Base58Version)
	miningKey, _ := committeePubKey.GetMiningKey(common.BlsConsensus)
	miningKeyStr := base58.Base58Check{}.Encode(miningKey, common.Base58Version)

	fmt.Println("Incognito Private Key: ", privateKeyStr)
	fmt.Println("Incognito Payment Address: ", paymentAddrStr)
	fmt.Println("Incognito Viewing Key: ", readOnlyKeyStr)

	fmt.Println("Private Seed:         ", privateSeedStr)
	fmt.Println("BLS public key:       ", committeePubKey.GetMiningKeyBase58(common.BlsConsensus))
	fmt.Println("Bridge public key:    ", committeePubKey.GetMiningKeyBase58(common.BridgeConsensus))
	fmt.Println("Mining Public Key (BLS+DSA): ", miningKeyStr)
	fmt.Println("Committee Public Key: ", committeeKeyStr)
	return nil
}

func TestGenerateFullKeyFromPrivateKey(t *testing.T) {
	privateKey := "112t8ro4JyjNxs1JtGt4HG9s39wY9QDz61H8tXuo28Ufb9HE9Pshqc8pdChjAs8BXEzkam3PaJc7yHfmYJVsc5NG47eTijME4RqfS9JcR1u9"
	GenerateFullKeyFromPrivateKey(privateKey)
}

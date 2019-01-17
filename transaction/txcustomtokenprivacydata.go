package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/wallet"
	"strconv"
)

type TxTokenPrivacyData struct {
	TxNormal       Tx          // used for privacy functionality
	PropertyID     common.Hash // = hash of TxCustomTokenprivacy data
	PropertyName   string
	PropertySymbol string

	Type     int    // action type
	Mintable bool   // default false
	Amount   uint64 // init amount
}

func (self TxTokenPrivacyData) String() string {
	record := self.PropertyName
	record += self.PropertySymbol
	record += fmt.Sprintf("%d", self.Amount)
	if self.TxNormal.Proof != nil {
		for _, out := range self.TxNormal.Proof.OutputCoins {
			record += string(out.CoinDetails.PublicKey.Compress())
			record += strconv.FormatUint(out.CoinDetails.Value, 10)
		}
		for _, in := range self.TxNormal.Proof.InputCoins {
			if in.CoinDetails.PublicKey != nil {
				record += string(in.CoinDetails.PublicKey.Compress())
			}
			if in.CoinDetails.Value > 0 {
				record += strconv.FormatUint(in.CoinDetails.Value, 10)
			}
		}
	}
	return record
}

func (self TxTokenPrivacyData) JSONString() string {
	data, err := json.MarshalIndent(self, common.EmptyString, "\t")
	if err != nil {
		Logger.log.Error(err)
		return common.EmptyString
	}
	return string(data)
}

// Hash - return hash of custom token data, be used as Token ID
func (self TxTokenPrivacyData) Hash() (*common.Hash, error) {
	hash := common.DoubleHashH([]byte(self.String()))
	return &hash, nil
}

// CustomTokenParamTx - use for rpc request json body
type CustomTokenPrivacyParamTx struct {
	PropertyID     string                 `json:"TokenID"`
	PropertyName   string                 `json:"TokenName"`
	PropertySymbol string                 `json:"TokenSymbol"`
	Amount         uint64                 `json:"TokenAmount"`
	TokenTxType    int                    `json:"TokenTxType"`
	Receiver       []*privacy.PaymentInfo `json:"TokenReceiver"`
	TokenInput     []*privacy.InputCoin   `json:"TokenInput"`
}

// CreateCustomTokenReceiverArray - parse data frm rpc request to create a list vout for preparing to create a custom token tx
// data interface is a map[paymentt-address]{transferring-amount}
func CreateCustomTokenPrivacyReceiverArray(data interface{}) ([]*privacy.PaymentInfo, int64) {
	result := []*privacy.PaymentInfo{}
	voutsAmount := int64(0)
	receivers := data.(map[string]interface{})
	for key, value := range receivers {
		keyWallet, _ := wallet.Base58CheckDeserialize(key)
		keySet := keyWallet.KeySet
		temp := &privacy.PaymentInfo{
			PaymentAddress: keySet.PaymentAddress,
			Amount:         uint64(value.(float64)),
		}
		result = append(result, temp)
		voutsAmount += int64(temp.Amount)
	}
	return result, voutsAmount
}

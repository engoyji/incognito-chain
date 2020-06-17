package instruction

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"strings"
)

type SwapInst struct {
	chainID     string
	chainType   string
	inCommitee  []incognitokey.CommitteePublicKey
	outCommitee []incognitokey.CommitteePublicKey
}

func NewBeaconSwapInst() *SwapInst {
	return &SwapInst{
		chainID:   "-1",
		chainType: "beacon",
	}
}

func NewShardSwapInst(shardID byte) *SwapInst {
	return &SwapInst{
		chainID:   fmt.Sprintf("%v", shardID),
		chainType: "shard",
	}
}

func (s *SwapInst) ImportFromString(str string) error {
	strSplit := strings.Split(str, " ")
	if strSplit[0] != "swap" {
		return errors.New("Not swap instruction")
	}

	return nil
}

func (s SwapInst) GetType() uint {
	return I_SWAP
}

func (s SwapInst) ToString() []string {
	swapInstString := []string{"swap"}
	inCommitteeString, _ := incognitokey.CommitteeKeyListToString(s.inCommitee)
	swapInstString = append(swapInstString, strings.Join(inCommitteeString, ","))
	outCommitteeString, _ := incognitokey.CommitteeKeyListToString(s.outCommitee)
	swapInstString = append(swapInstString, strings.Join(outCommitteeString, ","))
	swapInstString = append(swapInstString, s.chainType)
	swapInstString = append(swapInstString, s.chainID)
	return swapInstString
}

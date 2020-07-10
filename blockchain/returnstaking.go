package blockchain

import (
	"fmt"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

type returnStakingInfo struct {
	SwapoutPubKey string
	FunderAddress privacy.PaymentAddress
	StakingTx     metadata.Transaction
}

func (blockchain *BlockChain) buildReturnStakingTxFromBeaconInstructions(
	curView *ShardBestState,
	beaconBlocks []*BeaconBlock,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (
	[]metadata.Transaction,
	[][]string,
	error,
) {
	responsedTxs := []metadata.Transaction{}
	responsedHashTxs := map[common.Hash]struct{}{} // capture hash of responsed tx
	errorInstructions := [][]string{}
	var err error
	mReturnStakingInfo, errIns, err := blockchain.getReturnStakingInfoFromBeaconInstructions(
		curView,
		beaconBlocks,
		shardID,
	)
	errorInstructions = append(errorInstructions, errIns...)
	if err != nil {
		return nil, nil, err
	}
	for txStakingHash, returnInfo := range mReturnStakingInfo {
		txReturn, returnAmount, err := blockchain.buildReturnStakingAmountTx(
			curView,
			&returnInfo,
			producerPrivateKey,
			shardID,
		)
		Logger.log.Debugf("Return Staking Amount %v for funder %+v of candidate %v, staking transaction hash %+v, shardID %+v, err: %v", returnAmount, returnInfo.FunderAddress.String(), returnInfo.SwapoutPubKey, txStakingHash.String(), shardID, err)
		if err != nil {
			return nil, nil, err
		}
		txReturnHash := *txReturn.Hash()
		if _, ok := responsedHashTxs[txReturnHash]; ok {
			err = errors.Errorf("Double tx return staking from instruction for tx staking %v, swapout pubkey %v", txStakingHash, returnInfo.SwapoutPubKey)
			return nil, nil, err
		}
		responsedHashTxs[txReturnHash] = struct{}{}
		responsedTxs = append(responsedTxs, txReturn)
	}
	return responsedTxs, errorInstructions, nil
}

func (blockchain *BlockChain) ValidateReturnStakingTxFromBeaconInstructions(
	curView *ShardBestState,
	beaconBlocks []*BeaconBlock,
	shardBlock *ShardBlock,
	shardID byte,
) error {
	mReturnStakingInfoGot := map[common.Hash]returnStakingInfo{}
	for _, tx := range shardBlock.Body.Transactions {
		if tx.GetMetadataType() == metadata.ReturnStakingMeta {
			txHash := tx.Hash()
			returnMeta, ok := tx.GetMetadata().(*metadata.ReturnStakingMetadata)
			if !ok {
				return errors.Errorf("Can not parse metadata of tx %v to ReturnStaking Metadata", tx.Hash().String())
			}
			if _, ok := mReturnStakingInfoGot[*txHash]; ok {
				return errors.Errorf("Double tx return staking from instruction for tx staking %v", returnMeta.TxID)
			}
			h, _ := common.Hash{}.NewHashFromStr(returnMeta.TxID)
			mReturnStakingInfoGot[*h] = returnStakingInfo{
				FunderAddress: returnMeta.StakerAddress,
			}
		}
	}
	mReturnStakingInfoWanted, _, err := blockchain.getReturnStakingInfoFromBeaconInstructions(
		curView,
		beaconBlocks,
		shardID,
	)
	if err != nil {
		return err
	}
	if len(mReturnStakingInfoGot) != len(mReturnStakingInfoWanted) {
		return errors.Errorf("List return staking tx of producer (len %v) and validator (len %v) not match", len(mReturnStakingInfoGot), len(mReturnStakingInfoWanted))
	}
	for txStakingHash, returnInfoWanted := range mReturnStakingInfoWanted {
		if returnInfoGot, ok := mReturnStakingInfoGot[txStakingHash]; ok {
			if returnInfoGot.FunderAddress.String() != returnInfoWanted.FunderAddress.String() {
				return errors.Errorf("Validator want to return for funder %v using tx staking %v but producer return for %v", returnInfoWanted.FunderAddress.String(), returnInfoWanted.StakingTx.Hash().String(), returnInfoGot.FunderAddress.String())
			}
			continue
		}
		return errors.Errorf("Validator want to return for funder %v using tx staking %v but producer dont do it", returnInfoWanted.FunderAddress.String(), returnInfoWanted.StakingTx.Hash().String())
	}
	return nil
}

func (blockchain *BlockChain) buildReturnStakingAmountTx(
	curView *ShardBestState,
	info *returnStakingInfo,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (
	metadata.Transaction,
	uint64,
	error,
) {
	txStakingHash := info.StakingTx.Hash().String()
	Logger.log.Infof("Return Staking Amount public key %+v, staking transaction hash %+v, shardID %+v", info.SwapoutPubKey, txStakingHash, shardID)
	returnStakingMeta := metadata.NewReturnStaking(
		txStakingHash,
		info.FunderAddress,
		metadata.ReturnStakingMeta,
	)
	returnStakingTx := new(transaction.Tx)
	stakeAmount := info.StakingTx.CalculateTxValue()
	err := returnStakingTx.InitTxSalary(
		stakeAmount,
		&info.FunderAddress,
		producerPrivateKey,
		curView.GetCopiedTransactionStateDB(),
		returnStakingMeta,
	)
	//modify the type of the salary transaction
	returnStakingTx.Type = common.TxReturnStakingType
	if err != nil {
		return nil, 0, NewBlockChainError(InitSalaryTransactionError, err)
	}
	return returnStakingTx, stakeAmount, nil
}

func (blockchain *BlockChain) getReturnStakingInfoFromBeaconInstructions(
	curView *ShardBestState,
	beaconBlocks []*BeaconBlock,
	shardID byte,
) (
	map[common.Hash]returnStakingInfo,
	[][]string,
	error,
) {
	res := map[common.Hash]returnStakingInfo{}
	beaconConsensusStateDB := &statedb.StateDB{}
	beaconConsensusRootHash := common.Hash{}
	errorInstructions := [][]string{}
	beaconView := blockchain.BeaconChain.GetFinalView().(*BeaconBestState)
	var err error
	for _, beaconBlock := range beaconBlocks {
		beaconConsensusStateDB = nil
		for _, l := range beaconBlock.Body.Instructions {
			if l[0] == SwapAction {
				if beaconConsensusStateDB == nil {
					beaconConsensusRootHash, err = blockchain.GetBeaconConsensusRootHash(beaconView, beaconBlock.GetHeight())
					if err != nil {
						return nil, nil, NewBlockChainError(ProcessSalaryInstructionsError, fmt.Errorf("Beacon Consensus Root Hash of Height %+v not found, error %+v", beaconBlock.GetHeight(), err))
					}
					beaconConsensusStateDB, err = statedb.NewWithPrefixTrie(beaconConsensusRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetBeaconChainDatabase()))
				}
				for _, outPublicKey := range strings.Split(l[2], ",") {
					if len(outPublicKey) == 0 {
						continue
					}
					stakerInfo, has, err := statedb.GetStakerInfo(beaconConsensusStateDB, outPublicKey)
					if err != nil {
						Logger.log.Error(err)
						continue
					}
					if !has || stakerInfo == nil {
						Logger.log.Error(errors.Errorf("Can not found information of this public key %v", outPublicKey))
						continue
					}
					// If autostaking or staker who not has tx staking, do nothing
					if stakerInfo.AutoStaking() || (stakerInfo.TxStakingID() == common.HashH([]byte{0})) {
						continue
					}
					if _, ok := res[stakerInfo.TxStakingID()]; ok {
						err = errors.Errorf("Dupdate return staking using tx staking %v", stakerInfo.TxStakingID().String())
						return nil, nil, err
					}
					blockHash, index, err := rawdbv2.GetTransactionByHash(blockchain.GetShardChainDatabase(shardID), stakerInfo.TxStakingID())
					if err != nil {
						continue
					}
					shardBlock, _, err := blockchain.GetShardBlockByHash(blockHash)
					if err != nil || shardBlock == nil {
						Logger.log.Error("ERROR", err, "NO Transaction in block with hash", blockHash, "and index", index, "contains", shardBlock.Body.Transactions[index])
						continue
					}
					txData := shardBlock.Body.Transactions[index]
					keyWallet, err := wallet.Base58CheckDeserialize(txData.GetMetadata().(*metadata.StakingMetadata).FunderPaymentAddress)
					if err != nil {
						Logger.log.Error("SA: cannot get payment address", txData.GetMetadata().(*metadata.StakingMetadata), shardID)
						errorInstructions = append(errorInstructions, l)
						continue
					}
					Logger.log.Info("SA: build salary tx", txData.GetMetadata().(*metadata.StakingMetadata).FunderPaymentAddress, shardID)
					paymentShardID := common.GetShardIDFromLastByte(keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1])
					if paymentShardID != shardID {
						err = NewBlockChainError(WrongShardIDError, fmt.Errorf("Staking Payment Address ShardID %+v, Not From Current Shard %+v", paymentShardID, shardID))
						errorInstructions = append(errorInstructions, l)
						Logger.log.Error(err)
						continue
					}
					res[stakerInfo.TxStakingID()] = returnStakingInfo{
						SwapoutPubKey: outPublicKey,
						FunderAddress: keyWallet.KeySet.PaymentAddress,
						StakingTx:     txData,
					}
				}
			}
		}
	}
	return res, errorInstructions, nil
}

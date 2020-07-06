package rpcserver

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

func (httpServer *HttpServer) handlePreloadRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {

	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) != 2 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Invalid params"))
	}

	shardID, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Shard ID component invalid"))
	}
	epoch, ok := arrayParams[1].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Epoch component invalid"))
	}
	filePath := httpServer.config.ChainParams.BackupDir
	if shardID == -1 || shardID == 255 {
		filePath += "/beacon"
		if float64(httpServer.config.BlockChain.GetBeaconBestState().Epoch) <= epoch {
			return nil, rpcservice.NewRPCError(rpcservice.EpochIsLessOrEqual, fmt.Errorf("Provider epoch is %v", epoch))
		}
	} else {
		filePath = filePath + "/shard" + strconv.Itoa(int(shardID))
	}

	//Get needed epoch to download
	dirs, err := ioutil.ReadDir(filePath)
	if err != nil {
		fmt.Println("[backup-database] {HttpServer.handlePreloadRequest()} err:", err)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInternalError, err)
	}

	if len(dirs) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidRequestError, errors.New("Request data is not available"))
	}

	latestEpoch := 0

	for _, file := range dirs {
		tempEpoch, err := strconv.Atoi(file.Name())
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInternalError, err)
		}
		if tempEpoch > latestEpoch {
			latestEpoch = tempEpoch
		}
	}

	filePath = filePath + "/" + strconv.Itoa(int(latestEpoch))
	file, err := openFile(filePath)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidRequestError, err)
	}

	return file, nil
}

//openFile ...
func openFile(filePath string) (*os.File, error) {
	//Check if file exists and open
	file, err := os.Open(filePath)
	// defer Openfile.Close()

	if err != nil {
		return nil, err
	}

	return file, nil
}

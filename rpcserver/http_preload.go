package rpcserver

import (
	"errors"
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

	filePath := "./data/full_node/backup"

	if shardID == -1 || shardID == 255 {
		filePath += "/beacon"
	} else {
		filePath = filePath + "/shard" + strconv.Itoa(int(shardID))
	}

	filePath = filePath + "/" + strconv.Itoa(int(epoch))

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

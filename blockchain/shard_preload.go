package blockchain

func (blockchain *BlockChain) PreloadShardChainData(shardID byte) error {
	err := blockchain.GetShardChainDatabase(shardID).Close()
	if err != nil {
		return err
	}
	err = preloadDatabase(int(shardID), blockchain.config.ChainParams.PreloadFromAddr, blockchain.config.ChainParams.PreloadDir, blockchain.config.ChainParams.DataDir)
	if err != nil {
		Logger.log.Error(err)
		//panic(err)
	}
	err = blockchain.GetShardChainDatabase(shardID).ReOpen()
	if err != nil {
		return err
	}
	return nil
}

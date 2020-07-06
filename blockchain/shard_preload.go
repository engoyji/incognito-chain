package blockchain

func (blockchain *BlockChain) PreloadShardChainData(shardID byte) error {
	// we will force preload whole shard database
	err := preloadDatabase(int(shardID), 0, blockchain.config.ChainParams.PreloadFromAddr, blockchain.config.ChainParams.PreloadDir, blockchain.config.ChainParams.DataDir, blockchain.GetShardChainDatabase(shardID))
	if err != nil {
		Logger.log.Error(err)
		//panic(err)
	}
	if err := blockchain.RestoreShardViews(shardID); err != nil {
		return err
	}
	return nil
}

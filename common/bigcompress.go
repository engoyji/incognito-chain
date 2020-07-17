package common

import (
	"github.com/0xBahamoot/go-bigcompressor"
)

var bigCompress bigcompressor.BigCompressor

const (
	maxPrecompressChunkSize int64 = 104857600
	maxDecompressBufferSize int64 = 104857600
)

func CompressDatabase(src string, dst string) error {
	bigCompress.CombineChunk = true
	bigCompress.MaxPrecompressChunkSize = maxPrecompressChunkSize
	err := bigCompress.Compress(src, dst)
	if err != nil {
		return err
	}
	return nil
}

func DecompressDatabaseBackup(src string, dst string) error {
	bigCompress.MaxDecompressBufferSize = maxDecompressBufferSize
	err := bigCompress.Decompress(src, dst)
	if err != nil {
		return err
	}
	return nil
}

package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/utility/httprequest"
)

//JsonRequest ...
type JsonRequest struct {
	Jsonrpc string      `json:"Jsonrpc"`
	Method  string      `json:"Method"`
	Params  interface{} `json:"Params"`
	Id      interface{} `json:"Id"`
}

type RPCError struct {
	Code       int    `json:"Code,omitempty"`
	Message    string `json:"Message,omitempty"`
	StackTrace string `json:"StackTrace"`

	err error `json:"Err"`
}

type JsonResponse struct {
	Id      *interface{}    `json:"Id"`
	Result  json.RawMessage `json:"Result"`
	Error   *RPCError       `json:"Error"`
	Params  interface{}     `json:"Params"`
	Method  string          `json:"Method"`
	Jsonrpc string          `json:"Jsonrpc"`
}

//preloadDatabase call to backuped database node ...
func preloadDatabase(chainID int, currentEpoch uint64, url string, preloadDir string, dataDir string, db incdb.Database) error {

	//Send a json http request to backup database node
	header := http.Header{}
	header.Set("Content-Type", "application/json")
	header.Set("bin_resp", "true")
	req := JsonRequest{
		Jsonrpc: "2.0",
		Method:  "preload",
		Params:  []float64{float64(chainID), float64(currentEpoch)},
		Id:      1,
	}

	bodyReq, err := json.Marshal(req)
	if err != nil {
		return err
	}
	fmt.Println("[backup-database] {preloadDatabase} sent", chainID)
	resp, err := httprequest.Send(url, "POST", header, bodyReq)
	if err != nil {
		fmt.Println("[backup-database] {preloadDatabase} send request err:", err)
		return err
	}

	//Receive response and prepare for handling

	if resp.StatusCode != http.StatusOK {
		return errors.New("Error in preloading data, start normal sync")
	}

	defer resp.Body.Close()

	if resp.Header.Get("Content-Type") == "application/json" {
		//Json response mean can't get file
		// Reason could be backup database is caught up
		// or error in sending http process

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		jsonRes := JsonResponse{}
		err = json.Unmarshal(body, &jsonRes)
		if err != nil {
			return err
		}

		if jsonRes.Error.Code != -1001 {
			if jsonRes.Error.Code == -12001 {
				fmt.Println("No need to preload")
				return nil
			}
			fmt.Println("pkg blockchain {preloadDatabase} jsonRes.Error.Code:", jsonRes.Error.Code)
			return err
		}
		return nil
	}

	//Receive binary file
	// Read and Uncompress it

	path := preloadDir
	if chainID == -1 || chainID == 255 {
		path += "/beacon"
		dataDir += "/beacon"
	} else {
		path += "/shard" + strconv.Itoa(chainID)
		dataDir += "/shard" + strconv.Itoa(chainID)
	}

	defer resp.Body.Close()

	//Remove all old data
	if err := os.RemoveAll(path); err != nil {
		panic(err)
	}
	//Create new data
	if err := os.MkdirAll(path, 0700); err != nil {
		panic(err)
	}

	file, err := os.Create(path + "/" + resp.Header.Get("File-Name"))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	db.Close()
	db.Clear()
	defer db.ReOpen()

	err = uncompress(path+"/"+resp.Header.Get("File-Name"), dataDir)
	if err != nil {
		return err
	}

	// if chainID == 0 {
	// 	fmt.Println(path+"/"+resp.Header.Get("File-Name"), dataDir)
	// panic(0)
	// }
	return nil
}

//Uncompress file from zip file
func uncompress(srcPath, desPath string) error {

	//uncompress write
	//Remove all old data
	if err := os.RemoveAll(srcPath); err != nil {
		panic(err)
	}
	//Create new data
	if err := os.MkdirAll(srcPath, 0700); err != nil {
		panic(err)
	}

	if err := os.RemoveAll(desPath); err != nil {
		panic(err)
	}
	//Create new data
	if err := os.MkdirAll(desPath, 0700); err != nil {
		panic(err)
	}

	err := common.DecompressDatabaseBackup(srcPath, desPath)
	if err != nil {
		return err
	}
	return nil
}

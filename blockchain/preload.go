package blockchain

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/utility/httprequest"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	Id      *interface{}         `json:"Id"`
	Result  json.RawMessage      `json:"Result"`
	Error   *RPCError 			 `json:"Error"`
	Params  interface{}          `json:"Params"`
	Method  string               `json:"Method"`
	Jsonrpc string               `json:"Jsonrpc"`
}

//preloadDatabase call to backuped database node ...
func preloadDatabase(shardID int, url string, preloadDir string) error {

	//Send a json http request to backup database node
	header := http.Header{}
	header.Set("Content-Type", "application/json")
	header.Set("bin_resp", "true")
	req := JsonRequest{
		Jsonrpc: "2.0",
		Method:  "preload",
		Params:  []int{shardID},
		Id:      1,
	}

	bodyReq, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := httprequest.Send(url, "POST", header, bodyReq)
	if err != nil {
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

		if jsonRes.Error.Code != -1001{
			fmt.Println("pkg blockchain {preloadDatabase} jsonRes.Error.Code:", jsonRes.Error.Code)
			return errors.New("Wrong status code from backup database node")
		}
		return errors.New("Data is not available from this node")
	} else {
		//Receive binary file
		// Read and Uncompress it

		//"./data/untar"
		path := preloadDir
		if shardID == -1 || shardID == 255 {
			path += "/beacon"
		} else {
			path += "/shard" + strconv.Itoa(shardID)
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
		io.Copy(file, resp.Body)

		err = Uncompress(path + "/" + resp.Header.Get("File-Name"))
		if err != nil {
			return err
		}
	}

	return nil
}

////PreloadDatabase ...
//func PreloadDatabase(shardID int, url string) error {
//	err := preloadDatabase(shardID, url)
//	if err != nil {
//		return err
//	}
//	return nil
//}

//Uncompress file from zip file
func Uncompress(path string) error {

	//uncompress write
	//Remove all old data
	fd, _ := os.Open(path)
	if err := os.RemoveAll(path); err != nil {
		panic(err)
	}
	//Create new data
	if err := os.MkdirAll(path, 0700); err != nil {
		panic(err)
	}

	if err := uncompress(fd, path); err != nil {
		return err
	}
	return nil
}

//uncompress ...
func uncompress(src io.Reader, dst string) error {
	// ungzip
	zr, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	// untar
	tr := tar.NewReader(zr)

	// uncompress each element
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}
		target := filepath.Join(dst, header.Name)

		// if no join is needed, replace with ToSlash:
		// target = filepath.ToSlash(header.Name)

		// check the type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it (with 0755 permission)
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0700); err != nil {
					return err
				}
			}
		// if it's a file create it (with same permission)
		case tar.TypeReg:
			fileToWrite, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			// copy over contents
			if _, err := io.Copy(fileToWrite, tr); err != nil {
				return err
			}
			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			fileToWrite.Close()
		}
	}

	return nil
}
package preload

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/utility/httprequest"
	"io/ioutil"
	"net/http"
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

//PreloadDatabase ...
func PreloadDatabase(shardID int, url string, epoch uint64) error {

	for {
		//Send a json http request to backup database node
		header := http.Header{}
		header.Set("Content-Type", "application/json")
		req := JsonRequest{
			Jsonrpc: "2.0",
			Method:  "preload",
			Params:  []int{shardID, int(epoch)},
			Id:      0,
		}

		fmt.Println(req)

		bodyReq, err := json.Marshal(req)
		if err != nil {
			return err
		}

		resp, err := httprequest.Send(url, "POST", header, bodyReq)
		if err != nil {
			continue
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
				return errors.New("Wrong status code from backup database node")
			}
			break
		} else {
			//Receive binary file

		}
	}

	return nil
}

////JsonRequest ...
//type JsonRequest struct {
//	Jsonrpc string      `json:"Jsonrpc"`
//	Method  string      `json:"Method"`
//	Params  interface{} `json:"Params"`
//	Id      interface{} `json:"Id"`
//}
//
//////JsonResponse ...
////type JsonResponse struct {
////	Id      *interface{}         `json:"Id"`
////	Result  json.RawMessage      `json:"Result"`
////	Error   *rpcservice.RPCError `json:"Error"`
////	Params  interface{}          `json:"Params"`
////	Method  string               `json:"Method"`
////	Jsonrpc string               `json:"Jsonrpc"`
////}



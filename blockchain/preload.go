package blockchain

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/incognitochain/incognito-chain/rpcserver"
	"github.com/incognitochain/incognito-chain/utility/httprequest"
)

//preloadDatabase ...
func preloadDatabase(shardID int, url string, epoch uint64) error {

	for {
		//Send a json http request to backup database node
		header := http.Header{}
		header.Set("Content-Type", "application/json")
		req := rpcserver.JsonRequest{
			Jsonrpc: "2.0",
			Method:  "preload",
			Params:  []int{shardID, int(epoch)},
			Id:      0,
		}

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

			//body, err := ioutil.ReadAll(resp.Body)
			//if err != nil {
			//	return err
			//}
			//
			//jsonRes := preload.JsonResponse{}
			//
			//err = json.Unmarshal(body, &jsonRes)
			//if err != nil {
			//	return err
			//}
			//
			//if jsonRes.Error.Code != -1001{
			//	return errors.New("Wrong status code from backup database node")
			//}
			break
		} else {
			//Receive binary file

		}
	}

	return nil
}

package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	//"bytes"
)

var logger = shim.NewLogger("WalxSupplierChaincode")

// WalxSupplierChaincode ... Chaincode Object
type WalxSupplierChaincode struct {
}

// OrderFulfilment ... Object for Order Fulfilment
type OrderFulfilment struct {
	OrderNumber    string    `json:"order_number"`
	Gtin           string    `json:"gtin"`
	Quantity       int       `json:"quantity"`
	SupplierName   string    `json:"supplier_name"`
	Status         string    `json:"status"`
	CurLocation    string    `json:"current_location"`
	NxtLocation    string    `json:"next_location"`
	NxtLocationEta time.Time `json:"next_location_eta"`
	Ownership      string    `json:"ownership"`
}

// Init ...Chaincode initialization method
func (t *WalxSupplierChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init")
	return shim.Success(nil)
}

// Invoke ...Chaincode invoke method
func (t *WalxSupplierChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "createfulfilment" {
		return t.createFulfilment(stub, args)
	} else if function == "query" {
		return t.query(stub, args)
	}

	return pb.Response{Status: 403, Message: "unknown function name"}
}

func (t *WalxSupplierChaincode) createFulfilment(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error("cannot get creator")
	}

	user, org := getCreator(creatorBytes)
	if org == "" {
		logger.Debug("Org is null")
		return shim.Error("cannot get Org")
	} else if org == "walx" && user != "" {

		if len(args) < 6 {
			return pb.Response{Status: 403, Message: "incorrect number of arguments"}
		}

		quantity, err := strconv.Atoi(args[2])
		if err != nil {
			return shim.Error("Invalid Quantity, expecting a integer value")
		}

		strNxtLocationEta := args[7]

		nxtLocationEta, err := time.Parse(time.RFC3339, strNxtLocationEta)
		if err != nil {
			return shim.Error("Invalid ETA format, expecting a date value")
		}

		orderFulfilmentObj := &OrderFulfilment{
			OrderNumber:    args[0],
			Gtin:           args[1],
			Quantity:       quantity,
			SupplierName:   args[3],
			Status:         args[4],
			CurLocation:    args[5],
			NxtLocation:    args[6],
			NxtLocationEta: nxtLocationEta,
			Ownership:      args[8]}

		jsonOrderFulfilmentObj, err := json.Marshal(orderFulfilmentObj)
		if err != nil {
			return shim.Error("Cannot create JSON object")
		}

		poKey := args[0]

		err = stub.PutState(poKey, jsonOrderFulfilmentObj)
		if err != nil {
			return shim.Error("cannot put state")
		}

		logger.Debug("OrderFulfilment Object added")

	}
	return shim.Success(nil)
}

func (t *WalxSupplierChaincode) updateFulfilment(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error("cannot get creator")
	}

	user, org := getCreator(creatorBytes)
	if org == "" {
		logger.Debug("Org is null")
		return shim.Error("cannot get Org")
	} else if org == "supplier" && user != "" {

		if len(args) < 6 {
			return pb.Response{Status: 403, Message: "incorrect number of arguments"}
		}

		orderNumber := args[0]

		byteOrderFulfilment, err := stub.GetState(orderNumber)
		if err != nil {
			return shim.Error("cannot get state")
		} else if byteOrderFulfilment == nil {
			return shim.Error("Cannot get insurance object")
		}

		var orderFulfilmentObj OrderFulfilment

	}
	return shim.Success(nil)
}

func (t *WalxSupplierChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if args[0] == "health" {
		logger.Info("Health status Ok")
		return shim.Success(nil)
	}

	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error("cannot get creator")
	}

	user, org := getCreator(creatorBytes)

	if org == "" {
		logger.Debug("Org is null")
	} else if (org == "walx" || org == "customer") && user != "" {

		if len(args) != 1 {
			return pb.Response{Status: 403, Message: "incorrect number of arguments"}
		}

		key := args[0]

		bytes, err := stub.GetState(key)
		if err != nil {
			return shim.Error("cannot get state")
		}
		return shim.Success(bytes)
	} else if org == "banker" {

		if len(args) != 2 {
			return pb.Response{Status: 403, Message: "incorrect number of arguments"}
		}

		key := args[0] + "@" + args[1]

		bytes, err := stub.GetState(key)
		if err != nil {
			return shim.Error("cannot get state")
		}
		return shim.Success(bytes)
	}

	return shim.Success(nil)
}

var getCreator = func(certificate []byte) (string, string) {
	data := certificate[strings.Index(string(certificate), "-----") : strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, _ := x509.ParseCertificate(block.Bytes)
	organization := cert.Issuer.Organization[0]
	commonName := cert.Subject.CommonName
	logger.Debug("commonName: " + commonName + ", organization: " + organization)

	organizationShort := strings.Split(organization, ".")[0]

	return commonName, organizationShort
}

func main() {
	err := shim.Start(new(WalxSupplierChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}

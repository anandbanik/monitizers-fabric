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

var logger = shim.NewLogger("WalxCustomerChaincode")

// WalxCustomerChaincode ... Chaincode Object
type WalxCustomerChaincode struct {
}

// PurchaseOrder ... Struct for Purchase Order
type PurchaseOrder struct {
	PoNumber       string    `json:"po_number"`
	PoDate         time.Time `json:"po_date"`
	Gtin           string    `json:"gtin"`
	Quantity       int       `json:"quantity"`
	Quality        int       `json:"quality"`
	Time           int       `json:"time"`
	Sustainability int       `json:"sustainability"`
	Cost           int       `json:"cost"`
	Status         string    `json:"status"`
}

// Init ...Chaincode initialization method
func (t *WalxCustomerChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init")
	return shim.Success(nil)
}

// Invoke ...Chaincode invoke method
func (t *WalxCustomerChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "placeorder" {
		return t.placeorder(stub, args)
	} else if function == "query" {
		return t.query(stub, args)
	} else if function == "acceptorder" {
		return t.acceptorder(stub, args)
	} else if function == "receivedorder" {
		return t.receivedorder(stub, args)
	}

	return pb.Response{Status: 403, Message: "unknown function name"}
}

func (t *WalxCustomerChaincode) acceptorder(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error("cannot get creator")
	}

	user, org := getCreator(creatorBytes)

	logger.Debug("User: " + user)

	if org == "wmtx" {

		if len(args) != 2 {
			return pb.Response{Status: 403, Message: "incorrect number of arguments"}
		}

		poNumber := args[0]

		purchaseOrderBytes, err := stub.GetState(poNumber)

		if err != nil {
			return shim.Error("cannot get state")
		} else if purchaseOrderBytes == nil {
			return shim.Error("Cannot get po object")
		}

		var purchaseOrderObj PurchaseOrder
		errUnmarshal := json.Unmarshal([]byte(purchaseOrderBytes), &purchaseOrderObj)
		if errUnmarshal != nil {
			return shim.Error("Cannot unmarshal Negotiate Object")
		}

		logger.Debug("Negotiate Object: " + string(purchaseOrderBytes))

		purchaseOrderObj.Status = args[1]

		purchaseOrderObjBytes, _ := json.Marshal(purchaseOrderObj)

		errNegotiate := stub.PutState(poNumber, purchaseOrderObjBytes)
		if errNegotiate != nil {
			return shim.Error("Error updating PurchaseOrder Object: " + err.Error())
		}

		logger.Info("Update sucessfull")
		return shim.Success(nil)
	}
	return shim.Error("Org is not WmtX")

}

func (t *WalxCustomerChaincode) receivedorder(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error("cannot get creator")
	}

	user, org := getCreator(creatorBytes)

	logger.Debug("User: " + user)

	if org == "customer" {

		if len(args) != 2 {
			return pb.Response{Status: 403, Message: "incorrect number of arguments"}
		}

		poNumber := args[0]

		purchaseOrderBytes, err := stub.GetState(poNumber)

		if err != nil {
			return shim.Error("cannot get state")
		} else if purchaseOrderBytes == nil {
			return shim.Error("Cannot get po object")
		}

		var purchaseOrderObj PurchaseOrder
		errUnmarshal := json.Unmarshal([]byte(purchaseOrderBytes), &purchaseOrderObj)
		if errUnmarshal != nil {
			return shim.Error("Cannot unmarshal Negotiate Object")
		}

		logger.Debug("Negotiate Object: " + string(purchaseOrderBytes))

		purchaseOrderObj.Status = args[1]

		purchaseOrderObjBytes, _ := json.Marshal(purchaseOrderObj)

		errNegotiate := stub.PutState(poNumber, purchaseOrderObjBytes)
		if errNegotiate != nil {
			return shim.Error("Error updating PurchaseOrder Object: " + err.Error())
		}

		logger.Info("Update sucessfull")
		return shim.Success(nil)
	}
	return shim.Error("Org is not Customer")

}

func (t *WalxCustomerChaincode) placeorder(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error("cannot get creator")
	}

	user, org := getCreator(creatorBytes)
	if org == "" {
		logger.Debug("Org is null")
		return shim.Error("cannot get Org")
	} else if org == "customer" && user != "" {

		if len(args) < 7 {
			return pb.Response{Status: 403, Message: "incorrect number of arguments"}
		}

		location, err := time.LoadLocation("America/Chicago")
		if err != nil {
			fmt.Println(err)
		}

		poDate := time.Now().In(location)

		quantity, err := strconv.Atoi(args[2])
		if err != nil {
			return shim.Error("Invalid Quantity, expecting a integer value")
		}

		quality, err := strconv.Atoi(args[3])
		if err != nil {
			return shim.Error("Invalid Quality, expecting a integer value")
		}

		timeFoctor, err := strconv.Atoi(args[4])
		if err != nil {
			return shim.Error("Invalid Time factor, expecting a integer value")
		}

		sustainability, err := strconv.Atoi(args[5])
		if err != nil {
			return shim.Error("Invalid Sustainability factor, expecting a integer value")
		}

		cost, err := strconv.Atoi(args[6])
		if err != nil {
			return shim.Error("Invalid Cost, expecting a integer value")
		}

		purchaseOrderObj := &PurchaseOrder{
			PoNumber:       args[0],
			PoDate:         poDate,
			Gtin:           args[1],
			Quantity:       quantity,
			Quality:        quality,
			Time:           timeFoctor,
			Sustainability: sustainability,
			Cost:           cost,
			Status:         "Applied"}

		jsonPurchaseOrderObj, err := json.Marshal(purchaseOrderObj)
		if err != nil {
			return shim.Error("Cannot create Json Object")
		}

		logger.Debug("Json Obj: " + string(jsonPurchaseOrderObj))

		poKey := args[0]

		err = stub.PutState(poKey, jsonPurchaseOrderObj)

		if err != nil {
			return shim.Error("cannot put state")
		}

		logger.Debug("PO Object added")

	} else {
		logger.Warning("User is null or Org is not customer")
		return shim.Error("User is null or Org is not customer")

	}
	return shim.Success(nil)
}

func (t *WalxCustomerChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

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
	} else if (org == "wmtx" || org == "customer") && user != "" {

		if len(args) != 1 {
			return pb.Response{Status: 403, Message: "incorrect number of arguments"}
		}

		key := args[0]

		bytes, err := stub.GetState(key)
		if err != nil {
			return shim.Error("cannot get state")
		}
		return shim.Success(bytes)
	} else {
		return shim.Error("Check Org or user is null")
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
	err := shim.Start(new(WalxCustomerChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}

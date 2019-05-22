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
	}

	return pb.Response{Status: 403, Message: "unknown function name"}
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

		if len(args) < 6 {
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

		purchaseOrderObj := &PurchaseOrder{
			PoNumber:       args[0],
			PoDate:         poDate,
			Gtin:           args[1],
			Quantity:       quantity,
			Quality:        quality,
			Time:           timeFoctor,
			Sustainability: sustainability,
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
	err := shim.Start(new(WalxCustomerChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}

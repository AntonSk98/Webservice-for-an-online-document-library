package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-zookeeper/zk"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

var connection *zk.Conn

func main() {
	//setUpLaunchDelay()
	connection = createZooKeeperConnection()
	createSubscription(os.Getenv("GSERVE_NAME"))
	http.HandleFunc("/library", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost {
			doPostSaveData(writer, request, "http://hbase:8080/se2:library/fakerow")
		} else {
			fmt.Fprint(writer, fetchALlDataFromHbase(100), "\n", "proudly served by "+ os.Getenv("GSERVE_NAME"))
		}
	})
	log.Fatal(http.ListenAndServe(":80", nil))
}

func createZooKeeperConnection() *zk.Conn {
	connection, _, err := zk.Connect([]string{"zookeeper:2181"}, time.Second*10)
	if err != nil {
		log.Fatal(err)
	}
	return connection
}

func checkServiceZnodeAndCreateIfNeeded() {
	hbase, _, _ := connection.Exists("/hbase")
	for !hbase {
		time.Sleep(time.Second)
		hbase, _, _ = connection.Exists("/hbase")
	}
	setUpLaunchDelay()
	if exist, _, _ := connection.Exists("/services"); !exist {
		//fmt.Println("/SERVICES DOES NOT EXIST TRY TO RECREATE....")
		if _, err := connection.Create("/services", []byte(""), 0, zk.WorldACL(zk.PermAll)); err != nil {
			log.Fatal(err)
		}
	}
}

func createSubscription(name string) {
	checkServiceZnodeAndCreateIfNeeded()
	if _, err := connection.Create("/services/"+name, []byte(name), zk.FlagEphemeral, zk.WorldACL(zk.PermAll)); err != nil {
		time.Sleep(time.Second)
	}
}

func setUpLaunchDelay() {
	rand.Seed(time.Now().UnixNano())
	min := 5
	max := 15
	value := rand.Intn(max - min + 1) + min
	time.Sleep(time.Second * time.Duration(value))
}


func doPostSaveData(writer http.ResponseWriter, request *http.Request, hbaseUrl string) {
	response := saveData(encodeDocument(getRequestBody(request)), hbaseUrl)
	fmt.Println(response)
	fmt.Print("PROUDLY SERVED BY "+os.Getenv("GSERVE_NAME"))
}

func getRequestBody(request *http.Request) RowsType {
	unencodedJSON, _ := ioutil.ReadAll(request.Body)
	var document RowsType
	err := json.Unmarshal(unencodedJSON, &document)
	if err != nil {
		log.Fatal(err)
	}
	return document
}

func encodeDocument(document RowsType) EncRowsType {
	var encrDocumnet EncRowsType
	encrDocumnet.Row = make([]EncRowType, len(document.Row))
	for i:=0; i<len(document.Row); i++ {
		encrDocumnet.Row[i] = document.Row[i].encode()
	}
	return encrDocumnet
}

func saveData(encrDocumnet EncRowsType, hbaseUrl string) *http.Response {
	encodedDocumentJson, _ := json.Marshal(encrDocumnet)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, hbaseUrl, bytes.NewBuffer(encodedDocumentJson))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	return resp
}

func fetchALlDataFromHbase(batchValue int) RowsType {
	client := &http.Client{}
	getAllDataRequest, err := http.NewRequest(http.MethodPut, "http://hbase:8080/se2:library/scanner/", bytes.NewBuffer([]byte("{\"batch\": \""+strconv.Itoa(batchValue)+"\"}")))
	if err != nil {
		log.Fatal(err)
	}
	getAllDataRequest.Header.Set("Accept", "text/plain")
	getAllDataRequest.Header.Set("Content-Type", "application/json")
	response, err := client.Do(getAllDataRequest)
	if err != nil {
		log.Fatal(err)
	}
	location := response.Header.Get("Location")

	req, err := http.NewRequest(http.MethodGet, location, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "application/json")
	getAllDataResponse, _ := client.Do(req)
	var encodedRowsFromHbase EncRowsType
	unencodedJSON, _ := ioutil.ReadAll(getAllDataResponse.Body)
	if len(unencodedJSON) == 0 {
		return RowsType{}
	}
	if err := json.Unmarshal(unencodedJSON, &encodedRowsFromHbase); err != nil {
		log.Fatal(err)
	}
	deleteScannerRequest, _ := http.NewRequest(http.MethodDelete, location, nil)
	_, err = client.Do(deleteScannerRequest)
	if err != nil {
		log.Fatal(err)
	}
	return decryptDocument(encodedRowsFromHbase)
}

func decryptDocument(document EncRowsType) RowsType {
	var decryptedDoc RowsType
	decryptedDoc.Row = make([]RowType, len(document.Row))
	for i:=0; i<len(document.Row); i++ {
		decodedRow, err := document.Row[i].decode()
		if err != nil {
			log.Fatal(decodedRow)
		}
		decryptedDoc.Row[i] = decodedRow
	}
	return decryptedDoc
}

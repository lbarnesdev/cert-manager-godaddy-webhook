package godaddyclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// GodaddyClient is a godaddy.com client interface for manipulating text records for acme letsencrypt verification with cert-manager
type GodaddyClient interface {
	GetTXTRecord(string) ([]TxtRecord, error)
	AddTXTRecord(string, string) error
	DeleteTXTRecord(string) error
}

type godaddyClient struct {
	sandbox       bool
	authorization string
	domainName    string
	url           string
	zone          string
	client        http.Client
}

// GetTXTRecord gets a TXT record from godaddy dns
func (gc godaddyClient) GetTXTRecord(name string) ([]TxtRecord, error) {

	requestURL := fmt.Sprintf("%v/v1/domains/%v/records/TXT/%v", gc.url, gc.domainName, name)

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", gc.authorization)

	resp, err := gc.client.Do(req)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var txtRecords []TxtRecord
	json.Unmarshal(body, &txtRecords)
	return txtRecords, nil
}

// AddTXTRecord adds a txt record to godaddy dns
func (gc godaddyClient) AddTXTRecord(key string, value string) error {
	txtRecords := []TxtRecord{
		{
			Data: value,
			Name: key,
			Type: "TXT",
		},
	}

	json, err := json.Marshal(txtRecords)
	if err != nil {
		return err
	}

	requestURL := fmt.Sprintf("%v/v1/domains/%v/records", gc.url, gc.domainName)
	req, err := http.NewRequest(http.MethodPatch, requestURL, bytes.NewBuffer(json))
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", gc.authorization)

	resp, err := gc.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

// DeleteTXTRecord takes a string argument specifying TXT record name to delete
func (gc godaddyClient) DeleteTXTRecord(name string) error {

	txtRecordsURL := fmt.Sprintf("%v/v1/domains/%v/records/TXT", gc.url, gc.domainName)

	req, err := http.NewRequest(http.MethodGet, txtRecordsURL, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", gc.authorization)

	resp, err := gc.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	println(string(body))

	var txtRecords []TxtRecord
	json.Unmarshal(body, &txtRecords)
	fmt.Printf("deleteTXTRecords: %v\n", txtRecords)

	recordFound := false
	recordIndex := -1
	for i := 0; i < len(txtRecords) && !recordFound; i++ {
		println("iter: %v", i)
		if txtRecords[i].Name == name {
			println("found @: iter %v", i)
			recordFound = true
			recordIndex = i
		}
	}
	println(recordIndex)

	if !recordFound {
		return nil
	}

	txtRecords = append(txtRecords[:recordIndex], txtRecords[recordIndex+1:]...)
	fmt.Printf("after removal: %v\n", txtRecords)
	json, err := json.Marshal(&txtRecords)

	req, err = http.NewRequest(http.MethodPut, txtRecordsURL, bytes.NewBuffer(json))
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", gc.authorization)

	resp, err = gc.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	println(resp.StatusCode)

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	println(string(body))

	return nil
}

// New is a method for gettings an instance of godaddyClient
func New(domainName string, sandbox bool, key string, secret string) GodaddyClient {
	var url string
	if sandbox {
		url = "https://api.ote-godaddy.com"
	} else {
		url = "https://api.godaddy.com"
	}

	return godaddyClient{
		client: http.Client{
			Timeout: time.Second * 20,
		},
		sandbox:       sandbox,
		url:           url,
		authorization: fmt.Sprintf("sso-key %v:%v", key, secret),
		domainName:    domainName,
	}
}

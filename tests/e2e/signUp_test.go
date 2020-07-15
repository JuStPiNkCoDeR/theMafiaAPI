package e2e

import (
	main "../.."
	"../../utils"
	"encoding/json"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
)

const url = "ws://localhost:8001/ws/secure"

type Request struct {
	ReqID string `json:"reqID"`
	Name  string `json:"name"`
	Data  string `json:"data"`
}

func convertData(event string, data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)

	if err != nil {
		return "", err
	}

	req := &Request{
		ReqID: "sign up test",
		Name:  event,
		Data:  string(jsonData),
	}

	jsonReq, err := json.Marshal(req)

	if err != nil {
		return "", err
	}

	return string(jsonReq), nil
}

func TestSignUp(t *testing.T) {
	// Create test server with the secure socket handler
	server := httptest.NewServer(http.HandlerFunc(main.WsSecureHandler))
	defer server.Close()

	// Make the socket
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Can't make websocket\n%v", err)
	}
	defer func() {
		if err := ws.Close(); err != nil {
			t.Fatalf("Error on closing the socket\n%v", err)
		}
	}()

	var (
		pemOAEP            string
		pemPSS             string
		request            string
		response           []byte
		incomingMessage    interface{}
		incomingMessageMap map[string]interface{}
		incomingObjectData map[string]interface{}
		incomingSimpleData string
	)

	// Generate RSA keys
	r := &utils.RSA{}

	if err := r.Init(); err != nil {
		t.Fatalf("Can't initialize RSA object\n%v", err)
	}

	pemOAEP, pemPSS, err = r.ExportKeys()
	if err != nil {
		t.Fatalf("Can't export keys\n%v", err)
	}

	// Trigger 'rsa:getServerKeys' event
	request, err = convertData("rsa:getServerKeys", "")
	if err != nil {
		t.Fatalf("Can't convert data for rsa:getServerKeys event\n%v", err)
	}

	if err := ws.WriteMessage(websocket.TextMessage, []byte(request)); err != nil {
		t.Fatalf("Can't send message on 'rsa:getServerKeys' event\n%v", err)
	}

	// Here we should receive JSON string object should be like  {OAEP: string, PSS: string}
	_, response, err = ws.ReadMessage()
	if err != nil {
		t.Fatalf("Can't read message\n%v", err)
	}

	if err = json.Unmarshal(response, &incomingMessage); err != nil {
		t.Fatalf("Can't unmarshal response from 'rsa:getServerKeys' event\n%v", err)
	}

	incomingMessageMap = incomingMessage.(map[string]interface{})
	incomingObjectData = incomingMessageMap["data"].(map[string]interface{})

	// Import OAEP key
	if oaep, ok := incomingObjectData["OAEP"]; ok {
		if err := r.ImportKey(oaep.(string), true); err != nil {
			t.Fatalf("Can't import OAEP key\nPEM: %s\n%v", oaep, err)
		}
	} else {
		t.Fatal("No OAEP key presented at response from 'rsa:getServerKeys' event")
	}

	// Import PSS key
	if pss, ok := incomingObjectData["PSS"]; ok {
		if err := r.ImportKey(pss.(string), false); err != nil {
			t.Fatalf("Can't import PSS key\nPEM: %s\n%v", pss, err)
		}
	} else {
		t.Fatal("No PSS key presented at response from 'rsa:getServerKeys' event")
	}

	// Trigger 'rsa:setClientKeys' event
	request, err = convertData("rsa:setClientKeys", map[string]string{
		"oaep": pemOAEP,
		"pss":  pemPSS,
	})
	if err != nil {
		t.Fatalf("Can't convert data for rsa:setClientKeys event\n%v", err)
	}

	if err := ws.WriteMessage(websocket.TextMessage, []byte(request)); err != nil {
		t.Fatalf("Can't send message on 'rsa:setClientKeys' event\n%v", err)
	}

	// Here we should receive string "YES" or "NO"
	_, response, err = ws.ReadMessage()
	if err != nil {
		t.Fatalf("Can't read message\n%v", err)
	}

	if err = json.Unmarshal(response, &incomingMessage); err != nil {
		t.Fatalf("Can't unmarshal response from 'rsa:setClientKeys' event\n%v", err)
	}

	incomingMessageMap = incomingMessage.(map[string]interface{})
	incomingSimpleData = incomingMessageMap["data"].(string)

	if incomingSimpleData == "NO" {
		t.Fatal("Server response is 'NO' for 'rsa:setClientKeys' event")
	}

	// Trigger 'rsa:signUp' event

}

package libs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/joho/godotenv"

)

func GetExtractInter() string {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	app_key := os.Getenv("APP_KEY")
	//client_id := os.Getenv("CLIENT_ID")
	client_secret := os.Getenv("CLIENT_SECRET")
	agencia := os.Getenv("AGENCIA")
	conta := os.Getenv("CONTA")
	url := "https://api-extratos.bb.com.br/extratos/v1"
	requestBody := fmt.Sprintf("{%s\"conta-corrente\"agencia\"%s\"conta\"%s?gw-dev-app-key=%s}", url, agencia, conta, app_key)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		fmt.Println("Error creating request:", err)
	}
	req.Header.Set("Bearer %s", client_secret)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		client := &http.Client{}
		resp, err := client.Do(req)
	}

}



package netclient

import (
	"fmt"
	"os"
	"testing"
)

func TestLogToFile(t *testing.T) {
	f, err := os.OpenFile("testdata/test.log", os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	if err != nil {
		fmt.Println("Error opening file")
		t.Fail()
	}

	SetDebug(true)
	SetLoggerFile(f)

	client := getClient()

	resp, err := client.R().Get("http://httpbin.org/get")

	if err != nil {
		fmt.Printf("\nError: %v", err)
	}
	fmt.Printf("\nResponse Status Code: %v", resp.StatusCode())
	fmt.Printf("\nResponse Status: %v", resp.Status())
	fmt.Printf("\nResponse Body: %v", resp)
	fmt.Printf("\nResponse Time: %v", resp.Time())
	fmt.Printf("\nResponse Received At: %v", resp.ReceivedAt())

}

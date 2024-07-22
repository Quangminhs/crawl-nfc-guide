package main

import (
	"crawl/nfc/model"
	"encoding/base64"
	"fmt"
	"github.com/go-resty/resty/v2"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	numberOfWorkers = 10
)

func main() {
	host := os.Getenv("host")
	queue := startQueue(host)
	var wg sync.WaitGroup
	for i := 1; i <= numberOfWorkers; i++ {
		go crawlURL(queue, fmt.Sprintf("%d", i), &wg)
	}

	time.Sleep(1 * time.Second)

	wg.Wait()
	fmt.Println("finish crawl")
}

func crawlURL(queue <-chan string, name string, wg *sync.WaitGroup) {
	for v := range queue {
		wg.Add(1)
		err := crawlData(v, wg)
		if err != nil {
			fmt.Printf("Save error: " + v)
		}
		fmt.Printf("Worker %s is crawling URL %s\n", name, v)
	}

	fmt.Printf("Worker %s done and exit\n", name)
}

func startQueue(host string) <-chan string {
	queue := make(chan string, 1000)

	go func() {
		branchModel := model.NewBranchModel()
		for brand, modelNames := range branchModel.GetListBranch() {
			for _, name := range modelNames {
				rawUrl := host + "api/get-device-data?brand=" + brand + "&name=" + url.QueryEscape(name)
				queue <- rawUrl
			}
		}
		close(queue)
	}()

	return queue
}

func crawlData(rawUrl string, wg *sync.WaitGroup) error {
	defer wg.Done()
	client := resty.New()
	//Thực hiện GET request đến một API endpoint
	resp, err := client.R().
		EnableTrace().
		Get(rawUrl)

	// Xử lý lỗi nếu có
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	jsonString := resp.String()
	phone, err := model.ConvertToPerson(jsonString)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	u, err := url.Parse(rawUrl)
	if err != nil {
		return err
	}

	v, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return err
	}

	brand := v.Get("brand")
	folder := "./" + brand
	err = ensureFolderExists(folder)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	str := strings.ReplaceAll(v.Get("name"), " ", "_")

	cErr := ConvertImage(phone.Data.Image, folder+"/"+brand+"_"+str+".png")
	if cErr != nil {
		fmt.Println("Error:", cErr)
		return err
	}

	return nil

}

func ensureFolderExists(folderPath string) error {
	// Kiểm tra xem thư mục tồn tại hay không
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		// Nếu thư mục không tồn tại, tạo mới nó
		err := os.Mkdir(folderPath, 0755)
		if err != nil {
			return fmt.Errorf("error creating directory: %v", err)
		}
		fmt.Printf("Folder '%s' created successfully.\n", folderPath)
	} else if err != nil {
		// Xảy ra lỗi khi kiểm tra thư mục tồn tại
		return fmt.Errorf("error checking directory: %v", err)
	} else {
		//fmt.Printf("Folder '%s' already exists.\n", folderPath)
	}
	return nil
}

func ConvertImage(image, outputFileName string) error {
	// Extract the actual Base64 encoded image data
	base64String := "data:image/png;base64," + image
	idx := strings.Index(base64String, ";base64,")
	if idx < 0 {
		return fmt.Errorf("invalid Base64 string format")
	}
	base64Data := base64String[idx+8:]

	// Decode Base64 data
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("error decoding Base64: %v", err)
	}

	// Write the decoded data to a file
	err = ioutil.WriteFile(outputFileName, imageData, 0644)
	if err != nil {
		return fmt.Errorf("error writing image to file: %v", err)
	}

	fmt.Printf("Image saved to %s\n", outputFileName)
	return nil
}

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
)

var accessToken = "AgAEA7qi4nN5AAWwBP2NThoJI09GrSb21qEX85s" // Insert access token here

var sizes = []int{10, 25, 50, 100, 200} // Sire array in MB

var client = &http.Client{}

var  folderName = "upload-test"

func main() {

	log.Println("Check folder " + folderName + " existing...")
	req, err := http.NewRequest("GET", "https://cloud-api.yandex.net/v1/disk/resources?path=/" + folderName, nil)

	if err != nil {
		log.Fatalln(err)
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "OAuth " + accessToken)


	res, err := client.Do(req)

	if err != nil {
		log.Fatalln(err)
		return
	}

	if res.StatusCode == 404 {
		log.Println(" Folder not exists")
		err := createFolder()

		if err != nil {
			log.Fatalln(err)
			return
		}
	} else {
		log.Println("Folder exists, continue...")
	}

	for _, size := range sizes {
		log.Println("Generate file in " + strconv.Itoa(size) + " megabytes")
			dataString := make([]byte, size*1024*1024)
			rand.Read(dataString)

		link, err := getUploadLink(size)

		if err != nil {
			log.Fatalln(err)
			return
		}

		log.Println("Link to upload file is: " + link)

		err = upload(link, dataString)

		if err != nil {
			log.Fatalln(err)
			return
		}
	}

	cleanTestFolder()

}

func getJsonData(value []byte) map[string]interface{} {
	var dat map[string]interface{}

	if err := json.Unmarshal(value, &dat); err != nil {
		panic(err)
	}

	return dat
}

func createFolder() error {
	path := url.QueryEscape("/" + folderName)
	req, err := http.NewRequest("PUT", "https://cloud-api.yandex.net/v1/disk/resources?path=" + path, nil)

	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "OAuth " + accessToken)

	log.Println("Creating folder " + folderName)

	res, err := client.Do(req)

	if err != nil {
		return err
	}

	if res.StatusCode != 201 {
		return errors.New("Incorrect code " + res.Status)
	}

	log.Println("Folder created, continue...")

	return nil
}

func getUploadLink(size int) (string, error) {
	path := url.QueryEscape("/" + folderName + "/" + strconv.Itoa(size) + "MB.zip")
	req, err := http.NewRequest("GET", "https://cloud-api.yandex.net/v1/disk/resources/upload?path=" + path, nil)

	if err != nil {
	return "nil", err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "OAuth " + accessToken)


	log.Println("Request upload link...")

	res, err := client.Do(req)

	if err != nil {
		return "nil", err
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	if err != nil {
		return "nil", err
	}

	data := getJsonData(body)

	return data["href"].(string), nil
}

func upload(link string, data []byte) error {
	req, err := http.NewRequest("PUT", link, bytes.NewBuffer(data))

	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "OAuth " + accessToken)

	log.Println("Start upload to link...")

	start := time.Now()
	res, err := client.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		return err
	}

	size := len(data)
	speed := float64(len(data))/elapsed.Seconds()


	log.Println("Upload ended with status " + res.Status)
	log.Printf("Total time: %s, size: %d in bytes, average speed: %s", elapsed, size, normalizedSpeed(speed))

	return nil
}

func normalizedSpeed(speed float64) string {
	prefix := []string{"Bytes/s", "kB/s", "mB/s", "gB/s"}
	count := 0

	for speed > 1024 {
		speed = speed / 1024
		count++
	}

	return fmt.Sprintf("%.2f%s", speed, prefix[count])
}

func cleanTestFolder() error {
	path := url.QueryEscape("/" + folderName)
	req, err := http.NewRequest("DELETE", "https://cloud-api.yandex.net/v1/disk/resources?path=" + path + "&permanently=true", nil)

	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "OAuth " + accessToken)

	log.Println("Cleaning folder " + folderName)

	res, err := client.Do(req)

	if err != nil {
		return err
	}

	if res.StatusCode != 204 {
		return errors.New("Incorrect code " + res.Status)
	}

	log.Println("Folder removed, test ended.")

	return nil
}
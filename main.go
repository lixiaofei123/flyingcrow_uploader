package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type File struct {
	Urls     []string `json:"urls"`
	FilePath string   `json:"filePath"`
	FileName string   `json:"fileName"`
}

type ServerResponse struct {
	Code   int    `json:"code"`
	Reason string `json:"reason"`
	File   File   `json:"data"`
}

func Upload(url, path, token string) (string, error) {

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	part, err := writer.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", err
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/file/upload", url), buf)
	if err != nil {
		return "", err
	}
	req.Header.Add("token", token)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("服务器返回状态码是 %d", resp.StatusCode)
	}

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil
	}

	serverResp := ServerResponse{}
	err = json.Unmarshal(respData, &serverResp)
	if err != nil {
		return "", err
	}

	if serverResp.Code == 200 {
		return FindFileUrlByPath(url, fmt.Sprintf("%s/%s", serverResp.File.FilePath, serverResp.File.FileName), token)
	}
	return "", errors.New(serverResp.Reason)

}

func FindFileUrlByPath(url, path, token string) (string, error) {
	client := &http.Client{}
	resp, err := client.Get(fmt.Sprintf("%s/api/file/file?path=%s&token=%s", url, path, token))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("服务器返回状态码是 %d", resp.StatusCode)
	}

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil
	}
	serverResp := ServerResponse{}
	err = json.Unmarshal(respData, &serverResp)
	if err != nil {
		return "", err
	}

	if serverResp.Code == 200 {
		return serverResp.File.Urls[len(serverResp.File.Urls)-1], nil
	}
	return "", errors.New(serverResp.Reason)
}

func main() {

	var serverUrl string
	var token string

	flag.StringVar(&serverUrl, "s", "", "图床地址")
	flag.StringVar(&token, "t", "", "上传用的token")

	flag.Parse()
	paths := flag.Args()

	var urls []string = make([]string, 0)
	for _, path := range paths {
		url, err := Upload(serverUrl, path, token)
		if err != nil {
			fmt.Printf("上传出错了，原因是%s\n", err.Error())
			os.Exit(-1)
		}
		urls = append(urls, url)
	}
	for _, url := range urls {
		fmt.Println(url)
	}
}

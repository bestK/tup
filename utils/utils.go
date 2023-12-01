package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func createPng(text string) (string, error) {
	// 创建一个新的 RGBA 图像
	img := image.NewRGBA(image.Rect(0, 0, 500, 100))

	// 使用黑色填充图像
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)

	// 在图像中添加文本
	addLabel(img, 10, 50, text)

	output := fmt.Sprintf("%s.png", text)
	// 将图像保存为 PNG 文件
	err := saveAsPNG(img, output)
	if err != nil {
		log.Fatalf("Failed to save image: %v", err)
		return "", err
	}

	return output, err
}

func addLabel(img *image.RGBA, x, y int, label string) {
	col := color.RGBA{255, 255, 255, 255} // white
	point := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)}
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}

func saveAsPNG(img *image.RGBA, savePath string) error {

	file, err := os.Create(filepath.Join("tmp", savePath))
	if err != nil {
		return err
	}
	defer file.Close()

	png.Encode(file, img)
	return nil
}

func Encode(inputFile string) (string, error) {
	baseName := filepath.Base(inputFile)
	fakeImage, err := createPng(baseName)
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "darwin":
	case "windows":
		cmdStr := fmt.Sprintf("copy /b %s + %s %s", fakeImage, inputFile, "out_"+fakeImage)
		cmd := exec.Command("cmd", "/C", cmdStr)

		currentDir, _ := filepath.Abs(".")
		// The directory in which to execute the command
		cmdDir := filepath.Join(currentDir, "tmp")
		cmd.Dir = cmdDir

		err := cmd.Run()
		if err != nil {
			fmt.Println("Error copying file:", err)
			os.Exit(1)
		}
	case "linux":
	}

	return fakeImage, nil
}

type TelegraphResponse struct {
	Src string `json:"src"`
}

type PartItem struct {
	Url  string `json:"url"`
	Size int64  `json:"size"`
}

func UploadPart(uploadUrl string, partPath string) (PartItem, error) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	file, err := os.Open(filepath.Join("tmp", partPath))
	if err != nil {
		return PartItem{}, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return PartItem{}, err
	}
	fileSize := fileInfo.Size()

	part, err := writer.CreateFormFile("file", filepath.Base(partPath))
	if err != nil {
		return PartItem{}, err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return PartItem{}, err
	}

	err = writer.Close()
	if err != nil {
		return PartItem{}, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", uploadUrl, payload)
	if err != nil {
		return PartItem{}, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		return PartItem{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return PartItem{}, err
	}

	var response []TelegraphResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return PartItem{}, err
	}

	u, err := url.Parse(uploadUrl)
	if err != nil {
		fmt.Println("Error Parse upload url:", err)
		os.Exit(1)
	}

	return PartItem{
		Url:  fmt.Sprintf("https://%s%s", u.Host, response[0].Src),
		Size: fileSize,
	}, nil
}

func WriteResultJson(jsonData []PartItem, currFileName string) {

	jsonString, err := json.Marshal(jsonData)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(currFileName+".json", jsonString, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

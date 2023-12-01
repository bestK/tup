package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/bestk/tup/utils"
)

var currFileName string

func main() {
	// Parse command-line arguments
	inputFile := flag.String("input", "", "Path to the file to upload,if > 5MB,split it ")
	size := flag.Int("size", 4, "Size of each part in MB")
	outDir := flag.String("output", "", "Directory to output the parts to")
	uploadUrl := flag.String("upload-url", "https://telegra.ph/upload", "The url to upload parts to")
	flag.Parse()

	// If outDir is empty, set it to the current directory
	if *outDir == "" {
		var err error
		*outDir, err = os.Getwd()
		if err != nil {
			fmt.Println("Error getting current directory:", err)
			os.Exit(1)
		}
	}

	// Get the file size
	stat, err := os.Stat(*inputFile)
	if err != nil {
		fmt.Println("Error getting file size:", err)
		os.Exit(1)
	}
	fileSize := stat.Size()

	// Check if the file size is smaller than the split size
	if fileSize < int64(*size*1024*1024) {
		fmt.Println("The file is smaller than the split size.")
		os.Exit(1)
	}

	// Open the input file
	file, err := os.Open(*inputFile)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	defer file.Close()

	// Create the output directory
	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Println("Error creating output directory:", err)
		os.Exit(1)
	}

	// Create the temporary directory
	if err := os.MkdirAll(filepath.Join(*outDir, "tmp"), 0755); err != nil {
		fmt.Println("Error creating temporary directory:", err)
		os.Exit(1)
	}

	// Get the base name of the input file
	currFileName := filepath.Base(*inputFile)

	// Split the file
	buf := make([]byte, *size*1024*1024)
	partNum := 0

	var outputPngs []string

	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("Error reading file:", err)
			os.Exit(1)
		}
		if n == 0 {
			break
		}

		partName := fmt.Sprintf("%s_part_%d", currFileName, partNum)
		// Write part to disk
		outPath := filepath.Join(*outDir, "tmp", partName)
		outFile, err := os.Create(outPath)
		if err != nil {
			fmt.Println("Error creating output file:", err)
			os.Exit(1)
		}
		if _, err := outFile.Write(buf[:n]); err != nil {
			outFile.Close()
			fmt.Println("Error writing output file:", err)
			os.Exit(1)
		}
		outFile.Close()
		partPng, err := utils.Encode(partName)
		if err != nil {
			fmt.Println("Error creating output file:", err)
			os.Exit(1)
		}

		outputPngs = append(outputPngs, partPng)

		partNum++
	}

	var result []utils.PartItem

	const numWorkers = 3

	// create a buffered channel to manage the workload
	tasks := make(chan string, len(outputPngs))

	// start the workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for partPath := range tasks {
				fmt.Println("Uploading part:", partPath)

				image, err := utils.UploadPart(*uploadUrl, fmt.Sprintf("out_%s", partPath))
				if err != nil {
					fmt.Println("Error uploading part:", err)
					continue
				}

				result = append(result, image)

				fmt.Printf("Uploaded url: %s Size: %d\n", image.Url, image.Size)
			}
		}()
	}

	// enqueue the tasks
	for _, partPath := range outputPngs {
		tasks <- partPath
	}

	// close the channel so the workers can exit
	close(tasks)

	// wait for the workers to finish
	wg.Wait()

	os.RemoveAll("tmp")

	utils.WriteResultJson(result, currFileName)
}

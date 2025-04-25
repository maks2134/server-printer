package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	port         = 9100
	outputFolder = "printed_documents"
)

func main() {
	if err := os.MkdirAll(outputFolder, 0755); err != nil {
		fmt.Printf("Error creating output folder: %v\n", err)
		return
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		return
	}
	defer listener.Close()

	fmt.Printf("Virtual printer server running on port %d\n", port)
	fmt.Printf("Documents will be saved to: %s\n", outputFolder)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		go handlePrintJob(conn)
	}
}

func handlePrintJob(conn net.Conn) {
	defer conn.Close()

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown_host"
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	tempFilename := filepath.Join(outputFolder, fmt.Sprintf("temp_print_%s.pdf", timestamp))
	finalFilename := filepath.Join(outputFolder, fmt.Sprintf("print_%s.pdf", timestamp))

	tempFile, err := os.Create(tempFilename)
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		return
	}
	defer tempFile.Close()

	writer := bufio.NewWriter(tempFile)
	written, err := io.Copy(writer, conn)
	if err != nil {
		fmt.Printf("Error saving document: %v\n", err)
		return
	}
	writer.Flush()

	fmt.Printf("Received document: %s (%d bytes)\n", tempFilename, written)

	err = addWatermark(tempFilename, finalFilename, hostname)
	if err != nil {
		fmt.Printf("Error adding watermark: %v\n", err)
		return
	}
	os.Remove(tempFilename)
	fmt.Printf("Document with watermark saved: %s\n", finalFilename)
}

func addWatermark(inputPath, outputPath, text string) error {
	cmd := exec.Command("gs",
		"-q",
		"-dBATCH",
		"-dNOPAUSE",
		"-sDEVICE=pdfwrite",
		"-sOutputFile="+outputPath,
		"-c", fmt.Sprintf(`<< /EndPage {
            0 eq {
                /Helvetica 24 selectfont
                0.5 setgray
                50 50 moveto
                (%s) show
            } if
            true
        } >> setpagedevice`, text),
		"-f", inputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ghostscript error: %v, output: %s", err, string(output))
	}
	return nil
}

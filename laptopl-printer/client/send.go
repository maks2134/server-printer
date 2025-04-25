package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	file, err := os.Open("test.pdf")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	conn, err := net.Dial("tcp", "localhost:9100")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	_, err = io.Copy(conn, file)
	if err != nil {
		panic(err)
	}

	fmt.Println("PDF успешно отправлен на виртуальный принтер!")
}

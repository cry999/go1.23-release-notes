package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net"
	"os"
)

func main() {
	echConfigListFile := flag.String("ech-config-list", "ech_config_list.data", "")
	addr := flag.String("addr", "localhost:4430", "")
	serverName := flag.String("server-name", "private.example.org", "")

	flag.Parse()

	configByte, err := os.ReadFile(*echConfigListFile)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	client := tls.Client(conn, &tls.Config{
		EncryptedClientHelloConfigList: configByte,
		ServerName:                     *serverName,
		InsecureSkipVerify:             true,
	})

	err = client.Handshake()
	log.Println("client handshake:", err)
}

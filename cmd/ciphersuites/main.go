package main

import (
	"crypto/tls"
	"log"
	"net"
)

func main() {
	// Start a TLS server

	l, err := net.Listen("tcp", "localhost:4430")
	if err != nil {
		log.Println(err)
		return
	}
	defer l.Close()

	go func() {
		for {
			err := func() error {
				conn, err := l.Accept()
				if err != nil {
					return err
				}
				defer conn.Close()

				server := tls.Server(conn, &tls.Config{
					InsecureSkipVerify: true,
					GetConfigForClient: func(chi *tls.ClientHelloInfo) (*tls.Config, error) {
						log.Println("Client CipherSuites:", len(chi.CipherSuites))
						for _, id := range chi.CipherSuites {
							log.Println(tls.CipherSuiteName(id))
						}
						return &tls.Config{}, nil
					},
				})
				defer server.Close()

				return server.Handshake()
			}()
			if err != nil {
				log.Println("server.Handshake:", err)
				return
			}
		}
	}()

	// Start a TLS client

	conn, err := net.Dial("tcp", "localhost:4430")
	if err != nil {
		log.Print("net.Dial:", err)
		return
	}
	defer conn.Close()

	client := tls.Client(conn, &tls.Config{
		MaxVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
	})
	defer client.Close()

	if err := client.Handshake(); err != nil {
		log.Print("client.Handshake:", err)
		return
	}
}

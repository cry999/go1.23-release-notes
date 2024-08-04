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
					MinVersion: tls.VersionTLS13,
					MaxVersion: tls.VersionTLS13,
					GetConfigForClient: func(chi *tls.ClientHelloInfo) (*tls.Config, error) {
						for _, v := range chi.SupportedVersions {
							log.Println("Version:", tls.VersionName(v))
						}
						for _, id := range chi.SupportedCurves {
							log.Println("Curve:", id.String())
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
		MinVersion:         tls.VersionTLS13, // TLS1.3 only
		MaxVersion:         tls.VersionTLS13,
		InsecureSkipVerify: true,
	})
	defer client.Close()

	if err := client.Handshake(); err != nil {
		log.Print("client.Handshake:", err)
		return
	}
}

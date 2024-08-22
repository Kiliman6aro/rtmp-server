package main

import (
	"encoding/binary"
	"io"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":1935")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Println("RTMP server started on port 1935")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Println("New connection:", conn.RemoteAddr())

	err := rtmpHandshake(conn)
	if err != nil {
		log.Println("RTMP handshake failed:", err)
		return
	}

	log.Println("RTMP handshake completed successfully")

	// Now we proceed to handling RTMP commands
	for {
		err := handleRTMPCommand(conn)
		if err != nil {
			log.Println("Error handling RTMP command:", err)
			return
		}
	}
}

func rtmpHandshake(conn net.Conn) error {
	// Reading C0 (1 byte)
	c0 := make([]byte, 1)
	if _, err := io.ReadFull(conn, c0); err != nil {
		return err
	}

	if c0[0] != 0x03 {
		return io.ErrUnexpectedEOF
	}

	// Reading C1 (1536 bytes)
	c1 := make([]byte, 1536)
	if _, err := io.ReadFull(conn, c1); err != nil {
		return err
	}

	// Sending S0 and S1
	if _, err := conn.Write(append([]byte{0x03}, c1...)); err != nil {
		return err
	}

	// Reading C2
	c2 := make([]byte, 1536)
	if _, err := io.ReadFull(conn, c2); err != nil {
		return err
	}

	// Sending S2
	if _, err := conn.Write(c2); err != nil {
		return err
	}

	return nil
}

func handleRTMPCommand(conn net.Conn) error {
	// Reading the first byte of the RTMP message header to understand its type
	header := make([]byte, 1)
	if _, err := io.ReadFull(conn, header); err != nil {
		return err
	}

	// Determining the header length based on the first byte
	headerLength := determineHeaderLength(header[0])

	// Reading the remaining bytes of the header
	fullHeader := make([]byte, headerLength-1)
	if _, err := io.ReadFull(conn, fullHeader); err != nil {
		return err
	}

	// Extracting the message length from the header
	messageLength := binary.BigEndian.Uint32(append([]byte{0x00}, fullHeader[4:7]...))

	// Reading the message data
	messageData := make([]byte, messageLength)
	if _, err := io.ReadFull(conn, messageData); err != nil {
		return err
	}

	// Example parsing of AMF messages
	return parseRTMPCommand(messageData)
}

func determineHeaderLength(firstByte byte) int {
	// Determining the header length based on the first byte
	// This is a simplified version; in reality, the length depends on the value of firstByte
	// For example:
	if firstByte&0xC0 == 0x00 {
		return 12 // Standard full RTMP header
	} else if firstByte&0xC0 == 0x40 {
		return 8 // Medium header
	} else if firstByte&0xC0 == 0x80 {
		return 4 // Minimal header
	} else {
		return 1 // Stream ID only
	}
}

func parseRTMPCommand(data []byte) error {
	// The implementation of AMF parsing should be here to get RTMP commands
	log.Printf("Received RTMP command: %x", data)
	return nil
}

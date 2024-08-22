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

	// Теперь переходим к обработке команд RTMP
	for {
		err := handleRTMPCommand(conn)
		if err != nil {
			log.Println("Error handling RTMP command:", err)
			return
		}
	}
}

func rtmpHandshake(conn net.Conn) error {
	// Читаем C0 (1 байт)
	c0 := make([]byte, 1)
	if _, err := io.ReadFull(conn, c0); err != nil {
		return err
	}

	if c0[0] != 0x03 {
		return io.ErrUnexpectedEOF
	}

	// Читаем C1 (1536 байт)
	c1 := make([]byte, 1536)
	if _, err := io.ReadFull(conn, c1); err != nil {
		return err
	}

	// Отправляем S0 и S1
	if _, err := conn.Write(append([]byte{0x03}, c1...)); err != nil {
		return err
	}

	// Читаем C2
	c2 := make([]byte, 1536)
	if _, err := io.ReadFull(conn, c2); err != nil {
		return err
	}

	// Отправляем S2
	if _, err := conn.Write(c2); err != nil {
		return err
	}

	return nil
}

func handleRTMPCommand(conn net.Conn) error {
	// Читаем первый байт заголовка RTMP-сообщения, чтобы понять его тип
	header := make([]byte, 1)
	if _, err := io.ReadFull(conn, header); err != nil {
		return err
	}

	// Определяем длину заголовка на основе первого байта
	headerLength := determineHeaderLength(header[0])

	// Читаем остальные байты заголовка
	fullHeader := make([]byte, headerLength-1)
	if _, err := io.ReadFull(conn, fullHeader); err != nil {
		return err
	}

	// Извлекаем длину сообщения из заголовка
	messageLength := binary.BigEndian.Uint32(append([]byte{0x00}, fullHeader[4:7]...))

	// Чтение данных сообщения
	messageData := make([]byte, messageLength)
	if _, err := io.ReadFull(conn, messageData); err != nil {
		return err
	}

	// Примерный разбор AMF сообщений
	return parseRTMPCommand(messageData)
}

func determineHeaderLength(firstByte byte) int {
	// Определяем длину заголовка на основе первого байта
	// Это упрощенная версия; в реальности длина зависит от значения firstByte
	// Например:
	if firstByte&0xC0 == 0x00 {
		return 12 // Стандартный полный заголовок RTMP
	} else if firstByte&0xC0 == 0x40 {
		return 8 // Средний заголовок
	} else if firstByte&0xC0 == 0x80 {
		return 4 // Минимальный заголовок
	} else {
		return 1 // Только идентификатор потока
	}
}

func parseRTMPCommand(data []byte) error {
	// Здесь должна быть реализация разбора AMF для получения команд RTMP
	log.Printf("Received RTMP command: %x", data)
	return nil
}

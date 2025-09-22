package gocamole

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

type Config struct {
	Protocol   string
	Hostname   string
	Port       string
	Username   string
	Password   string
	Security   string
	IgnoreCert string
	Screen     Screen
}

type Screen struct {
	Heigth string
	Width  string
	Dpi    string
}

func InitializeGuacdConnection(hostname string, port int) (net.Conn, error) {
	addr := fmt.Sprintf("%v:%d", hostname, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func StartHandshake(conn net.Conn, ws *websocket.Conn, r *bufio.Reader, config *Config) error {
	selectIns := selectHandshakeInstruction(config.Protocol)
	if _, err := conn.Write(selectIns.Bytes()); err != nil {
		return err
	}

	argsIns, err := receiveFromGuacd(r)
	if err != nil {
		return err
	}

	sizeIns := sizeHandshakeInstruction(config.Screen.Width, config.Screen.Heigth, config.Screen.Dpi)
	if _, err := conn.Write(sizeIns.Bytes()); err != nil {
		return err
	}

	audioIns := audioHandshakeInstruction([]string{"audio/ogg", "audio/L66,rate=44100,channels=2"})
	if _, err := conn.Write(audioIns.Bytes()); err != nil {
		return err
	}

	videoIns := videoHandshakeInstruction(nil)
	if _, err := conn.Write(videoIns.Bytes()); err != nil {
		return err
	}

	imageIns := imageHandshakeInstruction(nil)
	if _, err := conn.Write(imageIns.Bytes()); err != nil {
		return err
	}

	conIns := connectHandshakeInstruction(argsIns, config)
	if _, err := conn.Write(conIns.Bytes()); err != nil {
		return err
	}

	readyIns, err := receiveFromGuacd(r)
	if err != nil {
		return err
	}

	if readyIns.Opcode != "ready" {
		return fmt.Errorf("ready instruction not received from guacd")
	}

	return nil
}

func WriteToGuacd(conn net.Conn, ws *websocket.Conn) {
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("Error reading from websocket: ", err)
			return
		}

		// From guacamole.js
		// The Guacamole protocol instruction opcode reserved for arbitrary internal
		// use by tunnel implementations. The value of this opcode is guaranteed to be
		// the empty string (""). Tunnel implementations may use this opcode for any
		// purpose. It is currently used by the HTTP tunnel to mark the end of the HTTP
		// response, and by the WebSocket tunnel to transmit the tunnel UUID and send
		// connection stability test pings/responses.
		if bytes.HasPrefix(msg, []byte("0.")) {
			continue
		}

		if _, err := conn.Write(msg); err != nil {
			log.Println("Error writing to guacd: ", err)
			return
		}
	}
}

func ReadFromGuacd(conn net.Conn, ws *websocket.Conn, r *bufio.Reader) {
	for {
		ins, err := receiveFromGuacd(r)
		if err != nil {
			log.Println("Error reading from guacd: ", err)
			return
		}

		msg := []byte(ins.String())

		// From guacamole.js
		// The Guacamole protocol instruction opcode reserved for arbitrary internal
		// use by tunnel implementations. The value of this opcode is guaranteed to be
		// the empty string (""). Tunnel implementations may use this opcode for any
		// purpose. It is currently used by the HTTP tunnel to mark the end of the HTTP
		// response, and by the WebSocket tunnel to transmit the tunnel UUID and send
		// connection stability test pings/responses.
		if bytes.HasPrefix(msg, []byte("0.")) {
			continue
		}

		if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("Error writing to websocket: ", err)
			return
		}
	}
}

func receiveFromGuacd(r *bufio.Reader) (*Instruction, error) {
	response, err := r.ReadString(';')
	if err != nil {
		return nil, err
	}

	ins, err := parseInstruction(r, response)
	if err != nil {
		return nil, err
	}

	return ins, nil
}

func parseInstruction(reader *bufio.Reader, raw string) (*Instruction, error) {
	if raw[len(raw)-1] != ';' {
		return nil, fmt.Errorf("unexpected response from Guacd, does not end with ';': %s", raw)
	}

	raw = raw[:len(raw)-1]
	var opcode string
	var args []string

	for i, chunk := range strings.Split(raw, ",") {
		chunks := strings.SplitN(chunk, ".", 2)
		if strconv.Itoa(len(chunks[1])) != chunks[0] {
			// Handle situation where value contains ; character.
			// Observed in audio instruction where received len.value is 31.audio/L66;rate=44100,channels=2;
			fullResponseLength, _ := strconv.Atoi(chunks[0])
			fullResponse := chunks[1] + ";"

			for fullResponseLength < len(fullResponse) {
				response, err := reader.ReadString(';')
				if err != nil {
					return nil, err
				}

				fullResponse += response[:len(response)-1]
			}

			if fullResponseLength != len(fullResponse) {
				return nil, fmt.Errorf("corrupted instruction: %s.%s", chunks[0], fullResponse)
			}

			chunks[1] = fullResponse
		}

		if i == 0 {
			opcode = chunks[1]
		} else {
			args = append(args, chunks[1])
		}
	}

	ins := &Instruction{
		Opcode: opcode,
		Args:   args,
	}

	return ins, nil
}

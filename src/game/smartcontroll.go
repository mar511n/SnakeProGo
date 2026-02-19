package game

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

type SmartConnection struct {
	inputCode string
	mu        sync.Mutex
}

func (sc *SmartConnection) AppendInputCode(code string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if len(code) != 5 {
		return
	}
	newcode := ""
	for i := 0; i < 5; i++ {
		if code[i] == '1' || sc.inputCode[i] == '1' {
			newcode += "1"
		} else {
			newcode += "0"
		}
	}
	sc.inputCode = newcode
}

func (sc *SmartConnection) GetInputCode() string {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if len(sc.inputCode) != 5 {
		return "00000"
	}
	code := sc.inputCode
	sc.inputCode = "00000"
	return code
}

type SmartController struct {
	connections      []*SmartConnection // handles connections to clients
	connectedPlayers []string           // player names corresponding to connections
	OnNewConnection  func()             // callback to inform main game of new connection
	ChanAssignPlayer chan string        // channel to assign player name to new connections
	mu               sync.Mutex
}

func (sc *SmartController) GetConnectedPlayers() []string {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	// Return a copy to avoid race conditions
	conns := make([]string, len(sc.connectedPlayers))
	copy(conns, sc.connectedPlayers)
	return conns
}

func (sc *SmartController) ProcessInputs(inputframe *InputFrame, playerConfigs map[int]*PlayerConfig) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	for idx, name := range sc.connectedPlayers {
		if idx < len(sc.connections) {
			conn := sc.connections[idx]
			if conn != nil {
				inputCode := conn.GetInputCode()
				//if inputCode != "00000" {
				//	LogInfo("Processing input code for player %s: %s", name, inputCode)
				//}
				inputframe.ProcessSmartInput(inputCode, name, playerConfigs)
			}
		}
	}
}

func (sc *SmartController) StartServer(port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		LogError("Error starting smart controller server: %v", err)
		return
	}
	// Run accept loop in a goroutine so it doesn't block the game
	go func() {
		defer listener.Close()
		LogInfo("Smart Controller Server listening on port %d", port)
		for {
			conn, err := listener.Accept()
			if err != nil {
				LogError("Error accepting connection: %v", err)
				continue
			}
			go sc.handleClient(conn)
		}
	}()
}

func (sc *SmartController) handleClient(conn net.Conn) {
	defer conn.Close()

	if sc.OnNewConnection == nil {
		return
	}
	go sc.OnNewConnection()

	playername := <-sc.ChanAssignPlayer
	smartConn := &SmartConnection{inputCode: "00000"} // Init with neutral, length 5

	sc.mu.Lock()
	sc.connections = append(sc.connections, smartConn)
	sc.connectedPlayers = append(sc.connectedPlayers, playername)
	sc.mu.Unlock()

	defer func() {
		sc.mu.Lock()
		defer sc.mu.Unlock()
		// Find and remove the disconnected player
		for i, name := range sc.connectedPlayers {
			if name == playername {
				sc.connectedPlayers = append(sc.connectedPlayers[:i], sc.connectedPlayers[i+1:]...)
				if i < len(sc.connections) {
					sc.connections = append(sc.connections[:i], sc.connections[i+1:]...)
				}
				break
			}
		}
	}()

	reader := bufio.NewReader(conn)
	for {
		// Expecting line-delimited input codes like "10000\n" or "00100"
		line, err := reader.ReadString('\n')
		if err != nil {
			LogError("Error reading from smart controller connection: %v", err)
		}
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			smartConn.AppendInputCode(line)
		}
	}
}

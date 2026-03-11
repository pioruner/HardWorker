package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/simonvetter/modbus"
)

type ViskoRow struct {
	T1   float64 `json:"t1"`
	T2   float64 `json:"t2"`
	U1   float64 `json:"u1"`
	U2   float64 `json:"u2"`
	Temp float64 `json:"temp"`
}

type ViskoControls struct {
	Address string `json:"address"`
}

type ViskoSnapshot struct {
	Connected    bool       `json:"connected"`
	LastResponse string     `json:"lastResponse"`
	Address      string     `json:"address"`
	Rows         []ViskoRow `json:"rows"`
	CursorIndex  int        `json:"cursorIndex"`
	CurT1        string     `json:"curT1"`
	CurT2        string     `json:"curT2"`
	CurU1        string     `json:"curU1"`
	CurU2        string     `json:"curU2"`
	CurTemp      string     `json:"curTemp"`
	CurCmd       string     `json:"curCmd"`
	SelT1        string     `json:"selT1"`
	SelT2        string     `json:"selT2"`
	SelU1        string     `json:"selU1"`
	SelU2        string     `json:"selU2"`
	SelTemp      string     `json:"selTemp"`
}

type ViskoService struct {
	id string

	mu   sync.RWMutex
	logs *LogBuffer

	address      string
	connected    bool
	lastResponse string
	rows         []ViskoRow
	cursorIndex  int
	curT1        string
	curT2        string
	curU1        string
	curU2        string
	curTemp      string
	curCmd       string
	selT1        string
	selT2        string
	selU1        string
	selU2        string
	selTemp      string
	lastCmd      uint16

	conn *modbus.ModbusClient
}

type persistedViskoState struct {
	Address     string `json:"address"`
	CursorIndex int    `json:"cursorIndex"`
}

func NewViskoService(id string) *ViskoService {
	s := &ViskoService{
		id:           id,
		address:      "192.168.0.200:502",
		lastResponse: "Ожидание подключения к вискозиметру...",
		rows:         make([]ViskoRow, 0, 256),
		cursorIndex:  0,
		curT1:        "0",
		curT2:        "0",
		curU1:        "0.00",
		curU2:        "0.00",
		curTemp:      "0.0",
		curCmd:       "0",
		selT1:        "0",
		selT2:        "0",
		selU1:        "0.00",
		selU2:        "0.00",
		selTemp:      "0.0",
		logs:         NewLogBuffer(500),
	}
	s.loadState()
	s.logInfo("VISCO service initialized")
	return s
}

func (s *ViskoService) Start(ctx context.Context) {
	go s.connectionLoop(ctx)
}

func (s *ViskoService) Shutdown() {
	s.saveState()
	s.mu.Lock()
	if s.conn != nil {
		_ = s.conn.Close()
		s.conn = nil
	}
	s.connected = false
	s.lastResponse = "Stopped"
	s.mu.Unlock()
	s.logInfo("VISCO service shutdown complete")
}

func (s *ViskoService) ApplyControls(in ViskoControls) {
	address := strings.TrimSpace(in.Address)
	if address == "" {
		return
	}

	s.mu.Lock()
	oldAddress := s.address
	s.address = address
	conn := s.conn
	s.mu.Unlock()

	if oldAddress != address {
		s.logInfo("Address changed to " + address)
		if conn != nil {
			_ = conn.Close()
		}
		s.saveState()
	}
}

func (s *ViskoService) SetCursorIndex(index int) {
	s.mu.Lock()
	if len(s.rows) == 0 {
		s.cursorIndex = 0
		s.mu.Unlock()
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(s.rows) {
		index = len(s.rows) - 1
	}
	s.cursorIndex = index
	s.updateSelectedLocked()
	s.mu.Unlock()
	s.saveState()
}

func (s *ViskoService) ClearRows() {
	s.mu.Lock()
	s.rows = s.rows[:0]
	s.cursorIndex = 0
	s.selT1, s.selT2, s.selU1, s.selU2, s.selTemp = "0", "0", "0.00", "0.00", "0.0"
	s.mu.Unlock()
	s.saveState()
	s.logInfo("Rows cleared")
}

func (s *ViskoService) statePath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "HardWorker")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, s.id+".json"), nil
}

func (s *ViskoService) saveState() {
	path, err := s.statePath()
	if err != nil {
		return
	}
	s.mu.RLock()
	state := persistedViskoState{
		Address:     s.address,
		CursorIndex: s.cursorIndex,
	}
	s.mu.RUnlock()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0o644)
}

func (s *ViskoService) loadState() {
	path, err := s.statePath()
	if err != nil {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var state persistedViskoState
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}

	address := strings.TrimSpace(state.Address)
	if address != "" {
		s.address = address
	}
	if state.CursorIndex >= 0 {
		s.cursorIndex = state.CursorIndex
	}
}

func (s *ViskoService) ExportCSV(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("empty path")
	}
	if !strings.HasSuffix(strings.ToLower(path), ".csv") {
		path += ".csv"
	}

	s.mu.RLock()
	rows := append([]ViskoRow(nil), s.rows...)
	s.mu.RUnlock()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"T1", "T2", "U1", "U2", "Temp"}); err != nil {
		return err
	}

	for _, r := range rows {
		record := []string{
			fmt.Sprintf("%.1f", r.T1),
			fmt.Sprintf("%.1f", r.T2),
			fmt.Sprintf("%.2f", r.U1),
			fmt.Sprintf("%.2f", r.U2),
			fmt.Sprintf("%.1f", r.Temp),
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}
	if err := w.Error(); err != nil {
		return err
	}
	return nil
}

func (s *ViskoService) GetSnapshot() ViskoSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return ViskoSnapshot{
		Connected:    s.connected,
		LastResponse: s.lastResponse,
		Address:      s.address,
		Rows:         append([]ViskoRow(nil), s.rows...),
		CursorIndex:  s.cursorIndex,
		CurT1:        s.curT1,
		CurT2:        s.curT2,
		CurU1:        s.curU1,
		CurU2:        s.curU2,
		CurTemp:      s.curTemp,
		CurCmd:       s.curCmd,
		SelT1:        s.selT1,
		SelT2:        s.selT2,
		SelU1:        s.selU1,
		SelU2:        s.selU2,
		SelTemp:      s.selTemp,
	}
}

func (s *ViskoService) GetLogs() []LogEntry {
	return s.logs.Snapshot()
}

func (s *ViskoService) connectionLoop(ctx context.Context) {
	retry := time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		address := s.getAddress()
		s.logInfo("Connecting to " + address)
		conn, err := modbus.NewClient(&modbus.ClientConfiguration{
			URL:     "tcp://" + address,
			Timeout: time.Second,
		})
		if err != nil {
			s.setDisconnected(err.Error())
			time.Sleep(retry)
			continue
		}

		if err := conn.Open(); err != nil {
			s.setDisconnected(err.Error())
			time.Sleep(retry)
			continue
		}

		s.mu.Lock()
		s.conn = conn
		s.connected = true
		s.lastResponse = "Connected"
		s.mu.Unlock()
		s.logInfo("Connected to " + address)

		err = s.sessionLoop(ctx, conn)
		_ = conn.Close()

		s.mu.Lock()
		if s.conn == conn {
			s.conn = nil
		}
		s.connected = false
		if err != nil {
			s.lastResponse = "Disconnected: " + err.Error()
			s.logWarn("Disconnected: " + err.Error())
		} else {
			s.lastResponse = "Disconnected"
			s.logInfo("Disconnected")
		}
		s.mu.Unlock()

		time.Sleep(retry)
	}
}

func (s *ViskoService) sessionLoop(ctx context.Context, conn *modbus.ModbusClient) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.readCycle(conn); err != nil {
				return err
			}
		}
	}
}

func (s *ViskoService) readCycle(conn *modbus.ModbusClient) error {
	conn.SetEncoding(modbus.BIG_ENDIAN, modbus.LOW_WORD_FIRST)

	t1, err := conn.ReadRegister(16384, modbus.HOLDING_REGISTER)
	if err != nil {
		s.logError("Read T1 failed: " + err.Error())
		return err
	}
	t2, err := conn.ReadRegister(16385, modbus.HOLDING_REGISTER)
	if err != nil {
		s.logError("Read T2 failed: " + err.Error())
		return err
	}
	temp, err := conn.ReadFloat32(16386, modbus.HOLDING_REGISTER)
	if err != nil {
		s.logError("Read Temp failed: " + err.Error())
		return err
	}
	u1, err := conn.ReadFloat32(16388, modbus.HOLDING_REGISTER)
	if err != nil {
		s.logError("Read U1 failed: " + err.Error())
		return err
	}
	u2, err := conn.ReadFloat32(16390, modbus.HOLDING_REGISTER)
	if err != nil {
		s.logError("Read U2 failed: " + err.Error())
		return err
	}
	cmd, err := conn.ReadRegister(16392, modbus.HOLDING_REGISTER)
	if err != nil {
		s.logError("Read CMD failed: " + err.Error())
		return err
	}

	s.mu.Lock()
	s.curT1 = strconv.FormatUint(uint64(t1), 10)
	s.curT2 = strconv.FormatUint(uint64(t2), 10)
	s.curU1 = fmt.Sprintf("%.2f", u1)
	s.curU2 = fmt.Sprintf("%.2f", u2)
	s.curTemp = fmt.Sprintf("%.1f", temp)
	s.curCmd = strconv.FormatUint(uint64(cmd), 10)
	s.connected = true
	s.lastResponse = "Read ok"

	if s.lastCmd == 0 && cmd > 0 && t1 >= 100 && t2 >= 100 {
		s.rows = append(s.rows, ViskoRow{T1: float64(t1), T2: float64(t2), U1: float64(u1), U2: float64(u2), Temp: float64(temp)})
		if len(s.rows) > 5000 {
			s.rows = s.rows[len(s.rows)-5000:]
		}
		s.cursorIndex = len(s.rows) - 1
		s.updateSelectedLocked()
	}
	s.lastCmd = cmd
	s.mu.Unlock()
	return nil
}

func (s *ViskoService) updateSelectedLocked() {
	if len(s.rows) == 0 {
		s.selT1, s.selT2, s.selU1, s.selU2, s.selTemp = "0", "0", "0.00", "0.00", "0.0"
		return
	}
	if s.cursorIndex < 0 {
		s.cursorIndex = 0
	}
	if s.cursorIndex >= len(s.rows) {
		s.cursorIndex = len(s.rows) - 1
	}
	r := s.rows[s.cursorIndex]
	s.selT1 = fmt.Sprintf("%.1f", r.T1)
	s.selT2 = fmt.Sprintf("%.1f", r.T2)
	s.selU1 = fmt.Sprintf("%.2f", r.U1)
	s.selU2 = fmt.Sprintf("%.2f", r.U2)
	s.selTemp = fmt.Sprintf("%.1f", r.Temp)
}

func (s *ViskoService) setDisconnected(msg string) {
	s.mu.Lock()
	s.connected = false
	s.lastResponse = msg
	s.mu.Unlock()
	s.logWarn("Connection failed: " + msg)
}

func (s *ViskoService) getAddress() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.address
}

func (s *ViskoService) logInfo(message string) {
	s.logs.Add(LogInfo, message)
}

func (s *ViskoService) logWarn(message string) {
	s.logs.Add(LogWarn, message)
}

func (s *ViskoService) logError(message string) {
	s.logs.Add(LogError, message)
}

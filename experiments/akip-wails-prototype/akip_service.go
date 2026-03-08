package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	timeScale = []float64{
		1e-6 * 1,
		1e-6 * 2,
		1e-6 * 5,
		1e-6 * 10,
		1e-6 * 20,
		1e-6 * 50,
		1e-6 * 100,
	}
	timeScaleS = []string{"1us", "2us", "5us", "10us", "20us", "50us", "100us"}
	baseOffset = []float64{7.6 * 1, 7.6 * 2, 7.6 * 5, 7.6 * 10, 7.6 * 20, 7.6 * 50, 7.6 * 100}
)

type AkipControls struct {
	Address    string     `json:"address"`
	TimeBase   int        `json:"timeBase"`
	HOffset    string     `json:"hOffset"`
	Reper      string     `json:"reper"`
	Square     string     `json:"square"`
	MinY       string     `json:"minY"`
	MinMove    string     `json:"minMove"`
	AutoSearch bool       `json:"autoSearch"`
	CursorMode string     `json:"cursorMode"`
	CursorPos  [3]float64 `json:"cursorPos"`
}

type AkipSnapshot struct {
	Connected    bool       `json:"connected"`
	LastResponse string     `json:"lastResponse"`
	Address      string     `json:"address"`
	TimeBase     int        `json:"timeBase"`
	HOffset      string     `json:"hOffset"`
	Reper        string     `json:"reper"`
	Square       string     `json:"square"`
	MinY         string     `json:"minY"`
	MinMove      string     `json:"minMove"`
	AutoSearch   bool       `json:"autoSearch"`
	CursorMode   string     `json:"cursorMode"`
	CursorPos    [3]float64 `json:"cursorPos"`
	X            []float64  `json:"x"`
	Y            []float64  `json:"y"`
	VSpeed       string     `json:"vSpeed"`
	VTime        string     `json:"vTime"`
	Volume       string     `json:"volume"`
	Registration bool       `json:"registration"`
}

type AkipService struct {
	id string

	mu   sync.RWMutex
	logs *LogBuffer

	address    string
	timeBase   int
	hOffset    string
	reper      string
	square     string
	minY       string
	minMove    string
	autoSearch bool
	cursorMode string
	cursorPos  [3]float64

	connected    bool
	lastResponse string
	x            []float64
	y            []float64
	vSpeed       string
	vTime        string
	volume       string
	volumeAbs    float64
	volumeRef    float64
	registration bool
	grpcAddress  string

	conn  net.Conn
	cmdCh chan string

	regFile   *os.File
	regWriter *csv.Writer
}

func NewAkipService(id string) *AkipService {
	s := &AkipService{
		id:          id,
		address:     "192.168.0.100:3000",
		timeBase:    3,
		hOffset:     "0",
		reper:       "25",
		square:      "10",
		minY:        "20",
		minMove:     "0.3",
		autoSearch:  false,
		cursorMode:  "start",
		cursorPos:   [3]float64{18, 34, 62},
		x:           []float64{0, 1, 2, 3},
		y:           []float64{1, 1, 1, 1},
		vSpeed:      "0.00",
		vTime:       "0.00",
		volume:      "0.00",
		volumeAbs:   0,
		volumeRef:   0,
		grpcAddress: ":50051",
		cmdCh:       make(chan string, 16),
		logs:        NewLogBuffer(500),
	}

	s.loadState()
	s.logInfo("AKIP service initialized")
	return s
}

func (s *AkipService) Start(ctx context.Context) {
	go s.connectionLoop(ctx)
	go s.grpcLoop(ctx)
}

func (s *AkipService) Shutdown() {
	s.logInfo("AKIP service shutdown started")
	s.StopRegistration()
	s.saveState()
	s.mu.Lock()
	if s.conn != nil {
		_ = s.conn.Close()
		s.conn = nil
	}
	s.mu.Unlock()
	s.logInfo("AKIP service shutdown complete")
}

func (s *AkipService) ApplyControls(in AkipControls) {
	address := strings.TrimSpace(in.Address)
	if address == "" {
		address = s.address
	}

	s.mu.Lock()
	oldAddress := s.address
	oldTimeBase := s.timeBase
	oldHOffset := s.hOffset
	oldAutoSearch := s.autoSearch

	s.address = address
	s.timeBase = clampTimeBase(in.TimeBase)
	s.hOffset = in.HOffset
	s.reper = in.Reper
	s.square = in.Square
	s.minY = in.MinY
	s.minMove = in.MinMove
	s.autoSearch = in.AutoSearch
	if in.CursorMode == "start" || in.CursorMode == "reper" || in.CursorMode == "front" {
		s.cursorMode = in.CursorMode
	}
	s.cursorPos = in.CursorPos
	conn := s.conn
	s.mu.Unlock()

	if oldAddress != address {
		s.logInfo("Address changed to " + address)
	}
	if oldTimeBase != s.timeBase {
		s.logInfo(fmt.Sprintf("Timebase changed to %s", timeScaleS[s.timeBase]))
	}
	if oldHOffset != s.hOffset {
		s.logDebug("Horizontal offset changed to " + s.hOffset)
	}
	if oldAutoSearch != s.autoSearch {
		if s.autoSearch {
			s.logInfo("Autosearch enabled")
		} else {
			s.logInfo("Autosearch disabled")
		}
	}

	if oldAddress != address && conn != nil {
		_ = conn.Close()
	}

	if oldTimeBase != s.timeBase {
		s.enqueueCommand(s.timeBaseCommand())
		s.enqueueCommand(s.offsetCommand())
	} else if oldHOffset != s.hOffset {
		s.enqueueCommand(s.offsetCommand())
	}

	s.calc()
	s.saveState()
}

func (s *AkipService) GetSnapshot() AkipSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return AkipSnapshot{
		Connected:    s.connected,
		LastResponse: s.lastResponse,
		Address:      s.address,
		TimeBase:     s.timeBase,
		HOffset:      s.hOffset,
		Reper:        s.reper,
		Square:       s.square,
		MinY:         s.minY,
		MinMove:      s.minMove,
		AutoSearch:   s.autoSearch,
		CursorMode:   s.cursorMode,
		CursorPos:    s.cursorPos,
		X:            append([]float64(nil), s.x...),
		Y:            append([]float64(nil), s.y...),
		VSpeed:       s.vSpeed,
		VTime:        s.vTime,
		Volume:       s.volume,
		Registration: s.registration,
	}
}

func (s *AkipService) GetLogs() []LogEntry {
	return s.logs.Snapshot()
}

func (s *AkipService) volumeLevel() float32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return float32(s.volumeAbs - s.volumeRef)
}

func (s *AkipService) ZeroVolumeReference() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.volumeRef = s.volumeAbs
	s.volume = "0.00"
	s.lastResponse = fmt.Sprintf("Опорный объем установлен: %.2f", s.volumeRef)
	s.logInfo(s.lastResponse)
}

func (s *AkipService) StartRegistration(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("empty path")
	}
	if !strings.HasSuffix(strings.ToLower(path), ".csv") {
		path += ".csv"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_ = s.stopRegistrationLocked()

	fileExists := false
	if _, err := os.Stat(path); err == nil {
		fileExists = true
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	w := csv.NewWriter(f)
	if !fileExists {
		_ = w.Write([]string{"Date-Time", "Volume ml"})
		w.Flush()
	}

	s.regFile = f
	s.regWriter = w
	s.registration = true
	s.logInfo("Registration started: " + path)
	return nil
}

func (s *AkipService) StopRegistration() {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.stopRegistrationLocked()
}

func (s *AkipService) stopRegistrationLocked() error {
	if s.regWriter != nil {
		s.regWriter.Flush()
		s.regWriter = nil
	}
	var err error
	if s.regFile != nil {
		err = s.regFile.Close()
		s.regFile = nil
	}
	s.registration = false
	s.logs.Add(LogInfo, "Registration stopped")
	return err
}

func (s *AkipService) connectionLoop(ctx context.Context) {
	retry := time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		address := s.getAddress()
		s.logInfo("Connecting to " + address)
		conn, err := net.DialTimeout("tcp", address, time.Second)
		if err != nil {
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

		s.enqueueCommand(":SDSLSCPI#")
		s.enqueueCommand(s.timeBaseCommand())
		s.enqueueCommand(s.offsetCommand())
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

func (s *AkipService) sessionLoop(ctx context.Context, conn net.Conn) error {
	ticker := time.NewTicker(200 * time.Millisecond)
	regTicker := time.NewTicker(time.Second)
	defer ticker.Stop()
	defer regTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.readWave(conn); err != nil {
				return err
			}
		case cmd := <-s.cmdCh:
			if err := s.sendCommand(conn, cmd); err != nil {
				return err
			}
		case <-regTicker.C:
			s.writeRegistrationRow()
		}
	}
}

func (s *AkipService) writeRegistrationRow() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.registration || s.regWriter == nil {
		return
	}
	_ = s.regWriter.Write([]string{
		time.Now().Format(time.DateTime),
		s.volume,
	})
	s.regWriter.Flush()
}

func (s *AkipService) setDisconnected(msg string) {
	s.mu.Lock()
	s.connected = false
	s.lastResponse = msg
	s.mu.Unlock()
	s.logWarn("Connection failed: " + msg)
}

func (s *AkipService) getAddress() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.address
}

func (s *AkipService) enqueueCommand(cmd string) {
	if strings.TrimSpace(cmd) == "" {
		return
	}
	select {
	case s.cmdCh <- cmd:
	default:
	}
}

func (s *AkipService) timeBaseCommand() string {
	s.mu.RLock()
	tb := clampTimeBase(s.timeBase)
	s.mu.RUnlock()
	return fmt.Sprintf(":TIMebase:SCALe %s", timeScaleS[tb])
}

func (s *AkipService) offsetCommand() string {
	s.mu.RLock()
	tb := clampTimeBase(s.timeBase)
	hOffset := s.hOffset
	s.mu.RUnlock()
	hoff, err := strconv.ParseFloat(hOffset, 64)
	if err != nil {
		return ""
	}
	value := (hoff + baseOffset[tb]) / (timeScale[tb] / 50.0)
	return fmt.Sprintf(":TIMebase:HOFFset %d", int(value*1e-6))
}

func (s *AkipService) sendCommand(conn net.Conn, cmd string) error {
	_ = conn.SetWriteDeadline(time.Now().Add(time.Second))
	if _, err := conn.Write([]byte(cmd)); err != nil {
		s.logError("Command send failed: " + err.Error())
		return err
	}

	s.mu.Lock()
	s.lastResponse = cmd
	s.mu.Unlock()
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (s *AkipService) readWave(conn net.Conn) error {
	_ = conn.SetWriteDeadline(time.Now().Add(time.Second))
	if _, err := conn.Write([]byte("STARTBIN")); err != nil {
		s.logError("STARTBIN write failed: " + err.Error())
		return err
	}

	header := make([]byte, 12)
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := io.ReadFull(conn, header); err != nil {
		s.logError("Wave header read failed: " + err.Error())
		return err
	}

	size := binary.LittleEndian.Uint16(header[0:2])
	if size == 0 || size > 65535 {
		s.logError(fmt.Sprintf("Invalid payload size: %d", size))
		return fmt.Errorf("invalid payload size: %d", size)
	}

	payload := make([]byte, size)
	if _, err := io.ReadFull(conn, payload); err != nil {
		s.logError("Wave payload read failed: " + err.Error())
		return err
	}

	packet := append(header, payload...)
	data, hoffs, nData, err := unpackChannel(packet)
	if err != nil {
		s.logError("Wave unpack failed: " + err.Error())
		return err
	}

	tb := s.getTimeBase()
	dt := timeScale[tb] * 15.2 / float64(nData) * 1000000
	y := u8ToFloat(data)
	x := make([]float64, nData)
	for i := -nData / 2; i < nData/2; i++ {
		x[i+nData/2] = (float64(i) * dt) + hoffs
	}

	s.mu.Lock()
	s.x = x
	s.y = y
	s.mu.Unlock()

	s.findPeakAndAdjust(hoffs)
	s.calc()
	return nil
}

func (s *AkipService) getTimeBase() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return clampTimeBase(s.timeBase)
}

func (s *AkipService) calc() {
	s.mu.Lock()
	defer s.mu.Unlock()
	c0 := s.cursorPos[0] / 2
	c1 := s.cursorPos[1] / 2
	c2 := s.cursorPos[2] / 2
	if c1-c0 == 0 {
		return
	}

	waveTime := c2 - c0
	repSize, _ := strconv.ParseFloat(s.reper, 64)
	rep := repSize / ((c1 - c0) / 1000000)
	speed := rep / 100
	sq, _ := strconv.ParseFloat(s.square, 64)
	volume := waveTime * rep * sq / 1000000
	delta := volume - s.volumeRef
	s.vTime = fmt.Sprintf("%.2f", waveTime)
	s.vSpeed = fmt.Sprintf("%.2f", speed)
	s.volumeAbs = volume
	s.volume = fmt.Sprintf("%.2f", delta)
}

func (s *AkipService) findPeakAndAdjust(hoffs float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.autoSearch || len(s.x) == 0 || len(s.y) == 0 {
		return
	}

	minX, maxX := -1, -1
	for i, x := range s.x {
		if x > s.cursorPos[2]-0.5 && x < s.cursorPos[2]+0.5 {
			if minX < 0 {
				minX = i
			}
			maxX = i
		}
	}
	if minX < 0 || maxX < 0 {
		return
	}

	maxY, maxIdx := -1.0, minX
	for i := minX; i < maxX; i++ {
		if s.y[i] > maxY {
			maxY = s.y[i]
			maxIdx = i
		}
	}

	offset := s.x[len(s.x)/2] - s.x[maxIdx]
	minR, _ := strconv.ParseFloat(s.minY, 64)
	minMove, _ := strconv.ParseFloat(s.minMove, 64)
	if maxY <= minR {
		return
	}
	s.cursorPos[2] = s.x[maxIdx]
	if math.Abs(offset) < minMove {
		offset = 0
	}
	if offset == 0 {
		return
	}

	tb := clampTimeBase(s.timeBase)
	s.hOffset = fmt.Sprintf("%.0f", hoffs-offset-baseOffset[tb])
	go s.enqueueCommand(s.offsetCommand())
}

func unpackChannel(packet []byte) ([]int8, float64, int, error) {
	size := binary.LittleEndian.Uint16(packet[0:2])
	dataBuf := packet[12:size]
	ch1 := bytes.Index(dataBuf, []byte{0x43, 0x48, 0x31})
	if ch1 == -1 {
		return nil, 0, 0, fmt.Errorf("channel CH1 not found")
	}

	nData := int(binary.LittleEndian.Uint16(dataBuf[ch1+15 : ch1+17]))
	hMove := binary.LittleEndian.Uint16(dataBuf[ch1+31 : ch1+33])
	hoffs := float64(math.Float32frombits(binary.LittleEndian.Uint32(dataBuf[ch1-12 : ch1-8])))
	shift := ch1 + 59
	wave := make([]int8, nData)
	for i := 0; i < nData; i++ {
		if shift+2 > len(dataBuf) {
			return nil, 0, 0, fmt.Errorf("wave buffer overflow")
		}
		wave[i] = int8(binary.LittleEndian.Uint16(dataBuf[shift:shift+2]) + hMove)
		shift += 2
	}

	return wave, hoffs, nData, nil
}

func u8ToFloat(data []int8) []float64 {
	out := make([]float64, len(data))
	for i, v := range data {
		out[i] = float64(v)
	}
	return out
}

func clampTimeBase(in int) int {
	if in < 0 {
		return 0
	}
	if in >= len(timeScale) {
		return len(timeScale) - 1
	}
	return in
}

type persistedState struct {
	Address    string     `json:"address"`
	TimeBase   int        `json:"timeBase"`
	HOffset    string     `json:"hOffset"`
	Reper      string     `json:"reper"`
	Square     string     `json:"square"`
	MinY       string     `json:"minY"`
	MinMove    string     `json:"minMove"`
	AutoSearch bool       `json:"autoSearch"`
	CursorMode string     `json:"cursorMode"`
	CursorPos  [3]float64 `json:"cursorPos"`
}

func (s *AkipService) statePath() (string, error) {
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

func (s *AkipService) saveState() {
	path, err := s.statePath()
	if err != nil {
		return
	}

	s.mu.RLock()
	state := persistedState{
		Address:    s.address,
		TimeBase:   s.timeBase,
		HOffset:    s.hOffset,
		Reper:      s.reper,
		Square:     s.square,
		MinY:       s.minY,
		MinMove:    s.minMove,
		AutoSearch: s.autoSearch,
		CursorMode: s.cursorMode,
		CursorPos:  s.cursorPos,
	}
	s.mu.RUnlock()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0o644)
}

func (s *AkipService) loadState() {
	path, err := s.statePath()
	if err != nil {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		s.logInfo("No saved state file, using defaults")
		return
	}

	var state persistedState
	if err := json.Unmarshal(data, &state); err != nil {
		s.logWarn("State parse failed, using defaults")
		return
	}

	s.address = strings.TrimSpace(state.Address)
	if s.address == "" {
		s.address = "192.168.0.100:3000"
	}
	s.timeBase = clampTimeBase(state.TimeBase)
	s.hOffset = state.HOffset
	s.reper = state.Reper
	s.square = state.Square
	s.minY = state.MinY
	s.minMove = state.MinMove
	s.autoSearch = false
	if state.CursorMode == "start" || state.CursorMode == "reper" || state.CursorMode == "front" {
		s.cursorMode = state.CursorMode
	}
	s.cursorPos = state.CursorPos
	s.logInfo("State loaded from " + path)
}

func (s *AkipService) logInfo(message string) {
	s.logs.Add(LogInfo, message)
}

func (s *AkipService) logWarn(message string) {
	s.logs.Add(LogWarn, message)
}

func (s *AkipService) logError(message string) {
	s.logs.Add(LogError, message)
}

func (s *AkipService) logDebug(message string) {
	s.logs.Add(LogDebug, message)
}

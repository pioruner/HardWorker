package setts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/AllenDang/giu"
	"github.com/pioruner/HardWorker.git/pkg/app"
)

type SettsUI struct {
	adr            string
	gRPCPort       string
	pollIntervalMs string
	autoConnect    bool
	startInTray    bool
	saveStatus     string
	onApply        func(SettsState)
}

type SettsState struct {
	AKIPAddress    string `json:"akip_ip_address"`
	GRPCPort       string `json:"grpc_port"`
	PollIntervalMs string `json:"poll_interval_ms"`
	AutoConnect    bool   `json:"auto_connect"`
	StartInTray    bool   `json:"start_in_tray"`
}

func DefaultState() SettsState {
	return SettsState{
		AKIPAddress:    "192.168.0.100:3000",
		GRPCPort:       ":50051",
		PollIntervalMs: "1000",
		AutoConnect:    true,
		StartInTray:    false,
	}
}

func LoadOrDefault() SettsState {
	s, err := LoadState()
	if err != nil {
		return DefaultState()
	}
	return s
}

func Init(onApply func(SettsState)) *SettsUI {
	def := DefaultState()
	ui := &SettsUI{
		adr:            def.AKIPAddress,
		gRPCPort:       def.GRPCPort,
		pollIntervalMs: def.PollIntervalMs,
		autoConnect:    def.AutoConnect,
		startInTray:    def.StartInTray,
		onApply:        onApply,
	}
	ui.Load()
	return ui
}

func appConfigPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(base, "HardWorker")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	return filepath.Join(dir, "settings.json"), nil
}

func (ui *SettsUI) ExportState() SettsState {
	return SettsState{
		AKIPAddress:    ui.adr,
		GRPCPort:       ui.gRPCPort,
		PollIntervalMs: ui.pollIntervalMs,
		AutoConnect:    ui.autoConnect,
		StartInTray:    ui.startInTray,
	}
}

func (ui *SettsUI) ImportState(s SettsState) {
	ui.adr = s.AKIPAddress
	ui.gRPCPort = s.GRPCPort
	ui.pollIntervalMs = s.PollIntervalMs
	ui.autoConnect = s.AutoConnect
	ui.startInTray = s.StartInTray
}

func SaveState(s SettsState) error {
	path, err := appConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func LoadState() (SettsState, error) {
	var state SettsState

	path, err := appConfigPath()
	if err != nil {
		return state, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return state, err
	}

	if err := json.Unmarshal(data, &state); err != nil {
		return state, err
	}

	return state, nil
}

func (ui *SettsUI) Save() error {
	return SaveState(ui.ExportState())
}

func (ui *SettsUI) Load() error {
	state, err := LoadState()
	if err != nil {
		return err
	}
	ui.ImportState(state)
	return nil
}

func (ui *SettsUI) UI() giu.Layout {
	return giu.Layout{
		giu.Column(
			giu.Row(
				giu.Label("AKIP адрес"),
				giu.InputText(&ui.adr).Size(260),
				giu.Dummy(20, 0),
				giu.Label("gRPC порт"),
				giu.InputText(&ui.gRPCPort).Size(120),
				giu.Dummy(20, 0),
				giu.Label("Опрос (мс)"),
				giu.InputText(&ui.pollIntervalMs).Size(100).Flags(giu.InputTextFlagsCharsDecimal),
			),
			giu.Separator(),
			giu.Row(
				giu.Checkbox("Автоподключение при старте", &ui.autoConnect),
				giu.Dummy(40, 0),
				giu.Checkbox("Запускать свернутым в трей", &ui.startInTray),
			),
			giu.Separator(),
			giu.Row(
				giu.Button("Сохранить").Size(150, 32).OnClick(func() {
					if err := ui.Save(); err != nil {
						ui.saveStatus = "Ошибка сохранения: " + err.Error()
						return
					}
					if ui.onApply != nil {
						ui.onApply(ui.ExportState())
					}
					ui.saveStatus = "Настройки сохранены: " + time.Now().Format("15:04:05")
				}),
				giu.Button("Перезагрузить").Size(150, 32).OnClick(func() {
					if err := ui.Load(); err != nil {
						ui.saveStatus = "Ошибка загрузки: " + err.Error()
						return
					}
					ui.saveStatus = "Настройки загружены: " + time.Now().Format("15:04:05")
				}),
				giu.Dummy(25, 0),
				giu.Button("Скрыть окно").Size(150, 32).OnClick(func() { app.Event <- app.EventToggleGUI }),
				giu.Button("Выход").Size(120, 32).OnClick(func() { app.Event <- app.EventQuit }),
			),
			giu.Separator(),
			giu.Label(ui.saveStatus),
			giu.Label("Порт gRPC применяется после перезапуска приложения."),
		),
	}
}

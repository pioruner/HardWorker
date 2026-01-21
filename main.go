package main

import (
	"context"
	_ "embed"
	"log"
	"time"

	"github.com/getlantern/systray"
)

// Импортируем иконку с помощью go:embed
//
//go:embed assets/icon.png
var iconData []byte

// Состояние приложения
type AppState struct {
	isConnected    bool
	isReconnecting bool
	ctx            context.Context
	cancel         context.CancelFunc
}

var appState *AppState

func main() {
	// Инициализируем состояние приложения
	ctx, cancel := context.WithCancel(context.Background())
	appState = &AppState{
		isConnected:    false,
		isReconnecting: false,
		ctx:            ctx,
		cancel:         cancel,
	}

	// Запускаем системный трей
	systray.Run(appState.onReady, appState.onExit)
}

func (a *AppState) onReady() {
	// Устанавливаем иконку
	if len(iconData) > 0 {
		systray.SetIcon(iconData)
	} else {
		log.Println("Иконка не найдена, будет использована стандартная")
	}

	// Заголовок и тултип
	systray.SetTitle("HardWorker")
	systray.SetTooltip("Управление устройствами")

	// Добавляем пункты меню
	mStatus := systray.AddMenuItem("Статус: Не подключено", "Текущий статус подключения")
	mStatus.Disable() // Только для отображения

	systray.AddSeparator()

	mReconnect := systray.AddMenuItem("Переподключить", "Переподключиться к устройствам")

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Выход", "Завершить программу")

	// Начинаем подключение при старте
	go a.connectToDevices()

	// Обработка событий в фоне
	go func() {
		for {
			select {
			case <-mReconnect.ClickedCh:
				a.reconnect()

			case <-mQuit.ClickedCh:
				systray.Quit()
				return

			case <-a.ctx.Done():
				return
			}
		}
	}()
}

func (a *AppState) connectToDevices() {
	if a.isReconnecting {
		return
	}

	a.isReconnecting = true
	defer func() { a.isReconnecting = false }()

	log.Println("Начинаю подключение к устройствам...")

	// Имитация процесса подключения
	for i := 1; i <= 5; i++ {
		select {
		case <-a.ctx.Done():
			log.Println("Подключение прервано")
			return
		default:
			// Имитация шага подключения
			time.Sleep(300 * time.Millisecond)
			log.Printf("Шаг подключения %d/5...", i)
		}
	}

	a.isConnected = true
	log.Println("Успешно подключено к устройствам")

	// Обновляем статус в трее
	systray.SetTitle("Device Manager ✓")
	systray.SetTooltip("Подключено к устройствам")
}

func (a *AppState) reconnect() {
	if a.isReconnecting {
		log.Println("Переподключение уже выполняется...")
		return
	}

	log.Println("Инициировано переподключение...")

	// Сбрасываем текущее подключение
	if a.isConnected {
		log.Println("Отключаю текущие соединения...")
		time.Sleep(500 * time.Millisecond)
		a.isConnected = false
	}

	// Обновляем статус
	systray.SetTitle("Device Manager")
	systray.SetTooltip("Переподключение...")

	// Запускаем новое подключение
	go a.connectToDevices()
}

func (a *AppState) onExit() {
	log.Println("Завершение работы приложения...")

	// Отменяем контекст для остановки всех горутин
	a.cancel()

	// Имитация очистки ресурсов
	log.Println("Закрытие соединений с устройствами...")
	time.Sleep(300 * time.Millisecond)

	log.Println("Приложение завершено")
}

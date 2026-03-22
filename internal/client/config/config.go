// Package config предоставляет конфигурацию для клиента GophKeeper.
// Параметры загружаются из (приоритет от высокого к низкому):
//   - Флаги командной строки
//   - Переменные окружения
//   - Значения по умолчанию
package config

import (
	"path/filepath"
	"time"
)

// ClientConfig содержит параметры конфигурации клиента.
type ClientConfig struct {
	ServerAddress  string        // адрес сервера (default: "localhost:50051")
	DataDir        string        // директория данных (default: "~/.gophkeeper")
	InsecureTLS    bool          // отключить проверку TLS (default: true)
	SyncInterval   time.Duration // интервал синхронизации (default: 1 мин)
	HistorySize    int           // размер истории команд (default: 500)
	HistoryFile    string        // путь к файлу истории
	CACertPath     string        // путь к CA сертификату
	ClientCertPath string        // путь к клиентскому сертификату (опционально)
	ClientKeyPath  string        // путь к ключу клиента (опционально)
}

// New создает новую конфигурацию из переменных окружения и флагов.
//
// Переменные окружения:
//
//	SERVER_ADDRESS, DATA_DIR, INSECURE_TLS, SYNC_INTERVAL, HISTORY_SIZE
//
// Флаги:
//
//	-server_address, -data_dir, -insecure_tls, -sync_interval, -history_size
func New() (*ClientConfig, error) {
	clientParameters, err := parseClientParameters()

	if err != nil {
		return nil, err
	}

	clientConfig := &ClientConfig{
		ServerAddress:  clientParameters.ServerAddress,
		DataDir:        clientParameters.DataDir,
		InsecureTLS:    clientParameters.InsecureTLS,
		SyncInterval:   time.Duration(clientParameters.SyncInterval) * time.Minute,
		HistorySize:    clientParameters.HistorySize,
		CACertPath:     clientParameters.CACertPath,
		ClientCertPath: clientParameters.ClientCertPath,
		ClientKeyPath:  clientParameters.ClientKeyPath,
	}

	clientConfig.HistoryFile = filepath.Join(clientParameters.DataDir, "history")

	return clientConfig, nil
}

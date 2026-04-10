// Package config предоставляет конфигурацию для клиента GophKeeper.
package config

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	serverAddress = "localhost:50051"
	insecureTLS   = true
	syncInterval  = 1 // минуты
	historySize   = 500
)

// ClientParameters содержит параметры конфигурации клиента.
type ClientParameters struct {
	ServerAddress  string
	DataDir        string
	InsecureTLS    bool
	SyncInterval   int
	HistorySize    int
	CACertPath     string
	ClientCertPath string
	ClientKeyPath  string
}

// parseClientParameters загружает параметры из .env, переменных окружения и флагов.
// Приоритет: флаги > переменные окружения > значения по умолчанию.
func parseClientParameters() (*ClientParameters, error) {
	_ = godotenv.Load()

	serverAddressParameter := serverAddressParameter()
	dataDirParameter := dataDirParameter()
	insecureTLSParameter := insecureTLSParameter()
	syncIntervalParameter := syncIntervalParameter()
	historySizeParameter := historySizeParameter()
	caCertPath := caCertPathParameter()
	clientCertPathParameter := clientCertPathParameter()
	clientKeyPathParameter := clientKeyPathParameter()

	flag.Parse()

	return &ClientParameters{
		ServerAddress:  serverAddressParameter,
		DataDir:        dataDirParameter,
		InsecureTLS:    insecureTLSParameter,
		SyncInterval:   syncIntervalParameter,
		HistorySize:    historySizeParameter,
		CACertPath:     caCertPath,
		ClientCertPath: clientCertPathParameter,
		ClientKeyPath:  clientKeyPathParameter,
	}, nil
}

func serverAddressParameter() string {
	val := serverAddress
	if env := os.Getenv("SERVER_ADDRESS"); env != "" {
		val = env
	}
	flag.StringVar(&val, "server_address", val, "server address")
	return val
}

func dataDirParameter() string {
	val := getDefaultDataDir()
	if env := os.Getenv("DATA_DIR"); env != "" {
		val = env
	}
	flag.StringVar(&val, "data_dir", val, "data directory")
	return val
}

func insecureTLSParameter() bool {
	val := insecureTLS
	if env := os.Getenv("INSECURE_TLS"); env != "" {
		val, _ = strconv.ParseBool(env)
	}
	flag.BoolVar(&val, "insecure_tls", val, "disable TLS verification")
	return val
}

func syncIntervalParameter() int {
	val := syncInterval
	if env := os.Getenv("SYNC_INTERVAL"); env != "" {
		if v, err := strconv.Atoi(env); err == nil {
			val = v
		}
	}
	flag.IntVar(&val, "sync_interval", val, "sync interval in minutes")
	return val
}

func historySizeParameter() int {
	val := historySize
	if env := os.Getenv("HISTORY_SIZE"); env != "" {
		if v, err := strconv.Atoi(env); err == nil {
			val = v
		}
	}
	flag.IntVar(&val, "history_size", val, "command history size")
	return val
}

// getDefaultDataDir возвращает путь по умолчанию: ~/.gophkeeper
func getDefaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".gophkeeper")
}

// Новые функции для TLS
func caCertPathParameter() string {
	caCertPath := ""
	if env, ok := os.LookupEnv("CA_CERT_PATH"); ok {
		caCertPath = env
	}
	flag.StringVar(&caCertPath, "ca_cert", caCertPath, "CA certificate path")
	return caCertPath
}

func clientCertPathParameter() string {
	clientCertPath := ""
	if env, ok := os.LookupEnv("CLIENT_CERT_PATH"); ok {
		clientCertPath = env
	}
	flag.StringVar(&clientCertPath, "client_cert", clientCertPath, "client certificate path")
	return clientCertPath
}

func clientKeyPathParameter() string {
	clientKeyPath := ""
	if env, ok := os.LookupEnv("CLIENT_KEY_PATH"); ok {
		clientKeyPath = env
	}
	flag.StringVar(&clientKeyPath, "client_key", clientKeyPath, "client key path")
	return clientKeyPath
}

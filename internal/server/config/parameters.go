// Package config содержит параметры конфигурации сервера сбора метрик.
package config

import (
	"flag"
	"os"
	"strconv"

	"github.com/Ko4etov/gophkeeper/internal/server/service/logger"
	"github.com/joho/godotenv"
)

var (
	address         = "50051"
	accessTokenTTL  = 15 // В минутах
	refreshTokenTTL = 24 // В Часах
)

// ServerParameters содержит все параметры конфигурации сервера.
type ServerParameters struct {
	Address         string // Адрес сервера
	DBAddress       string // Адрес базы данных
	HashKey         string // Ключ для хеширования
	CryptoKey       string // Файл с крипто ключом
	GRPCAddress     string
	JWTSecret       string
	AccessTokenTTL  int
	RefreshTokenTTL int
}

// parseServerParameters парсит параметры сервера из переменных окружения и флагов.
func parseServerParameters() *ServerParameters {
	if err := godotenv.Load(); err != nil {
		logger.Logger.Info(".env file not loaded: %v", err)
	}

	addressParameter := addressParameter()
	dbAddressParameter := dbAddressParameter()
	hashKeyParameter := hashKeyParameter()
	cryptoKeyParameter := cryptoKeyParameter()
	grpcAddressParameter := grpcAddressParameter()
	JWTSecretParameter := JWTSecretParameter()
	AccessTokenTTLParameter := accessTokenTTLParameter()
	RefreshTokenTTLParameter := refreshTokenTTLParameter()

	flag.Parse()

	parameters := &ServerParameters{
		Address:         addressParameter,
		DBAddress:       dbAddressParameter,
		HashKey:         hashKeyParameter,
		CryptoKey:       cryptoKeyParameter,
		GRPCAddress:     grpcAddressParameter,
		JWTSecret:       JWTSecretParameter,
		AccessTokenTTL:  AccessTokenTTLParameter,
		RefreshTokenTTL: RefreshTokenTTLParameter,
	}

	return parameters
}

// JWTSecretParameter
func JWTSecretParameter() string {
	jwtSecret := ""

	if env, ok := os.LookupEnv("JWT_SECRET"); ok {
		jwtSecret = env
	}

	flag.StringVar(&jwtSecret, "jwt_secret", jwtSecret, "Jwt secret key")

	return jwtSecret
}

// hashKeyParameter возвращает ключ для хеширования из переменных окружения или флагов.
func accessTokenTTLParameter() int {
	accessTokenTTL := accessTokenTTL

	if accessTokenTTLEnv, ok := os.LookupEnv("ACCESS_TOKEN_TTL"); ok {
		if val, err := strconv.Atoi(accessTokenTTLEnv); err == nil {
			accessTokenTTL = val
		}
	}
	flag.IntVar(&accessTokenTTL, "access_token_ttl", accessTokenTTL, "access token TTL in minutes")

	return accessTokenTTL
}

// refreshTokenTTLParameter возвращает ключ для хеширования из переменных окружения или флагов.
func refreshTokenTTLParameter() int {
	refreshTokenTTL := refreshTokenTTL

	if refreshTokenTTLEnv, ok := os.LookupEnv("REFRESH_TOKEN_TTL"); ok {
		if val, err := strconv.Atoi(refreshTokenTTLEnv); err == nil {
			refreshTokenTTL = val
		}
	}
	flag.IntVar(&refreshTokenTTL, "refresh_token_ttl", refreshTokenTTL, "refresh token TTL in hours")

	return refreshTokenTTL
}

// hashKeyParameter возвращает ключ для хеширования из переменных окружения или флагов.
func hashKeyParameter() string {
	hashKey := ""

	if env, ok := os.LookupEnv("HASH_KEY"); ok {
		hashKey = env
	}

	flag.StringVar(&hashKey, "hash_key", hashKey, "Hash key")

	return hashKey
}

// dbAddressParameter возвращает адрес базы данных из переменных окружения или флагов.
func dbAddressParameter() string {
	dbAddress := ""

	if env, ok := os.LookupEnv("DATABASE_DSN"); ok {
		dbAddress = env
	}

	flag.StringVar(&dbAddress, "db_address", dbAddress, "DB address")

	return dbAddress
}

// addressParameter возвращает адрес сервера из переменных окружения или флагов.
func addressParameter() string {
	address := address

	if env, ok := os.LookupEnv("ADDRESS"); ok {
		address = env
	}
	flag.StringVar(&address, "server_address", address, "Server address")

	return address
}

// cryptoKeyParameter возвращает путь до файла с крипто ключом из переменных окружения или флагов.
func cryptoKeyParameter() string {
	cryptoKey := ""

	if env, ok := os.LookupEnv("CRYPTO_KEY"); ok {
		cryptoKey = env
	}
	flag.StringVar(&cryptoKey, "crypto_key", cryptoKey, "Crypto key")

	return cryptoKey
}

func grpcAddressParameter() string {
	grpcAddr := ""

	if env, ok := os.LookupEnv("GRPC_ADDRESS"); ok {
		grpcAddr = env
	}

	flag.StringVar(&grpcAddr, "grpc_address", grpcAddr, "gRPC server address")

	return grpcAddr
}

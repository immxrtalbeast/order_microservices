module immxrtalbeast/order_microservices/saga-service

go 1.24.5

require immxrtalbeast/order_microservices/internal/pkg/tracing v0.0.0-00010101000000-000000000000

replace immxrtalbeast/order_microservices/internal/pkg/tracing => ../../internal/pkg/tracing

require (
	github.com/confluentinc/confluent-kafka-go v1.9.2
	github.com/fatih/color v1.18.0
	github.com/google/uuid v1.3.0
	github.com/ilyakaznacheev/cleanenv v1.5.0
	github.com/immxrtalbeast/order_kafka v0.0.0-20250917131923-8b079d832554
	github.com/joho/godotenv v1.5.1
	github.com/segmentio/kafka-go v0.4.49
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.30.1
)

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)

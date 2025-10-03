module auth-plugin

go 1.24.3

require (
	github.com/google/uuid v1.6.0
	golang.org/x/crypto v0.31.0
	gorm.io/gorm v1.31.0
)

replace leonidas/core => ../../

require leonidas/core v0.0.0-00010101000000-000000000000

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/text v0.21.0 // indirect
)

package model

type BucketCapacity struct {
	Login    int `yaml:"loginCapacity" env-default:"10" env:"LOGIN_CAP"`
	Password int `yaml:"pwdCapacity" env-default:"100" env:"PWD_CAP"`
	IP       int `yaml:"ipCapacity" env-default:"1000" env:"IP_CAP"`
}

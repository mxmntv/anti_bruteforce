package model

type BucketCapacity struct {
	Login    int `yaml:"N" env-default:"10" env:"N_CAP"`
	Password int `yaml:"M" env-default:"100" env:"M_CAP"`
	IP       int `yaml:"K" env-default:"1000" env:"K_CAP"`
}

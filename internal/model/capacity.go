package model

type BucketCapacity struct {
	Login    int `yaml:"n" env-default:"10" env:"N_CAP"`
	Password int `yaml:"m" env-default:"100" env:"M_CAP"`
	IP       int `yaml:"k" env-default:"1000" env:"K_CAP"`
}

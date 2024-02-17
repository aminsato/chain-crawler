package service

type Service interface {
	Run() (err error)
}

package service

// Service interface will be use by the service runner programm
// to run either a Creator and an Updator without any prior knowledge
// of the service being one or the other
type Service interface {
    Type() string
    Run()
}
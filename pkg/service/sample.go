package service

import (
	log "github.com/fidelity/theliv/pkg/log"
	"golang.org/x/net/context"
	err "github.com/fidelity/theliv/pkg/err"
)

type Address struct {
	Name     string
	Address string
}

func Hello(ctx context.Context) {
	a1 := GetSampleAddress()

	l := log.L(log.WithReqId(ctx))
	l.Info("Hello from Default Logger")

	s := log.S(log.WithReqId(ctx))
	s.Info("Hello from Suggared Logger")
	s.Infof("Address '%s': %s", a1.Name, a1.Address)

}

func GetSampleAddress() Address {
	return Address{
		"Sample1",
		"Address 1",
	}
}

func GetDefaultAddress() Address {
	return Address{
		"Default",
		"Deault address",
	}
}

func Error() (Address, error) {
	return Address{}, err.NewCommonError(err.COMMON, "Sample error")
}

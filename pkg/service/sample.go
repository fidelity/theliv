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

	l := log.LWithContext(ctx)
	l.Info("Hello from Default Logger")

	s := log.SWithContext(ctx)
	s.Info("Hello from Suggared Logger")
	s.Infof("Address '%s': %s", a1.Name, a1.Address)

	s1 := log.S()
	s1.Info("Test without context")
}

func Hello1(ctx context.Context) {
	log.S().Info("Test without context")
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

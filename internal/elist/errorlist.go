package elist

import (
	"strings"
)

type EList []error

func (el *EList) IsError() bool {
	if el == nil {
		return false
	}
	return len(*el) > 0
}

func (el *EList) Add(err error) {
	*el = append(*el, err)
}

func (el *EList) Error() string {
	v := strings.Builder{}
	for _, e := range *el {
		v.WriteString(e.Error())
		v.WriteRune('\n')
	}
	return v.String()
}

func (el *EList) ErrNil() error {
	if el.IsError() {
		return el
	}
	return nil
}

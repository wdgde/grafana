package v0alpha1

import "fmt"

func (s User) AuthID() string {
	return fmt.Sprintf("%d", s.Spec.InternalID)
}

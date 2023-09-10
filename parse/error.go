package parse

import "fmt"

type InvalidRoleError struct {
	Role string
}

func (r *InvalidRoleError) Error() string { return fmt.Sprintf("invalid role: %s", r.Role) }

package util

// Policy represents an AWS policy
type Policy struct {
	Version   string
	Statement []Statement
}

// Statement represents an AWS policy statement
type Statement struct {
	Effect   string
	Action   []string
	Resource string
}

// AssumeRolePolicy represents an AWS policy
type AssumeRolePolicy struct {
	Version   string
	Statement []RoleStatement
}

// RoleStatement represents an AWS policy statement
type RoleStatement struct {
	Sid       string
	Effect    string
	Action    string
	Principal map[string]string
}

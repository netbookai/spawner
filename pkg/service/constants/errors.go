package constants

import "fmt"

//bunch of common error message constants

const InvalidInstanceOrMachineType = "invalid instance, must provide valid instance by specifying MachineType or Instance as per provider specification"

var ErrInvalidCredentiualType = fmt.Errorf("invalid credentials type provided, must be one of ['%s', '%s', '%s']", CredAws, CredAzure, CredGitPat)

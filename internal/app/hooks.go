package app

const (
	preInstall  = "preInstall"
	postInstall = "postInstall"
	preUpgrade  = "preUpgrade"
	postUpgrade = "postUpgrade"
	preDelete   = "preDelete"
	postDelete  = "postDelete"
	test        = "test"
)

type hookCmd struct {
	Command
	Type string
}

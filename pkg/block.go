package pkg

import "github.com/hashicorp/hcl/v2"

type block interface {
	file() *hcl.File
	path() []string
	emitter() func(block Block) error
}

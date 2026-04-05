package sprites

import (
	"os"

	sprites "github.com/superfly/sprites-go"
)

var Client *sprites.Client

func init() {
	Client = sprites.New(os.Getenv("SPRITE_TOKEN"))
}

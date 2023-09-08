//go:build linux

package hooks

// Hooks to exececute host system commands

import (
	"regexp"

	"github.com/0xAX/notificator"

	"git.blob42.xyz/gomark/gosuki/parsing"
	"git.blob42.xyz/gomark/gosuki/tree"
)

// Hook that sends a system notification using notify-send (Linux).
// To enable notification a tag must be presetn in the name with the form: `#tag:notify`
// Requires notify-send to be installed
func NotifySend(node *tree.Node) error {
	// if node.Name has a tag of the form #tag:notify

	regex := regexp.MustCompile(parsing.ReNotify)
	
	if !regex.MatchString(node.Name) {
		return nil
	}

	notify := notificator.New(notificator.Options{
		AppName: "gosuki",
	})
	return notify.Push("new bookmark", node.URL, "", notificator.UR_NORMAL)
}

func init(){
	regHook(Hook{
		Name: "notify-send",
		Func: NotifySend,
	})
}





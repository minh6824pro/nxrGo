package config

import (
	"github.com/bwmarrin/snowflake"
	"log"
	"sync"
)

var (
	node *snowflake.Node
	once sync.Once
)

func GetSnowflakeNode() *snowflake.Node {
	once.Do(func() {
		var err error
		node, err = snowflake.NewNode(682) // Node ID (0-1023)
		if err != nil {
			log.Fatal(err)
		}
	})
	return node
}

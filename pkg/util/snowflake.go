package util

import (
	"fmt"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

// InitSnowflake 初始化雪花算法节点
// machineID: 机器 ID (0-1023)
func InitSnowflake(machineID int64) error {
	var err error
	node, err = snowflake.NewNode(machineID)
	if err != nil {
		return fmt.Errorf("failed to create snowflake node: %w", err)
	}
	return nil
}

// GenID 生成一个全局唯一的 int64 ID
func GenID() int64 {
	if node == nil {
		// 如果未手动初始化，默认使用节点 1
		_ = InitSnowflake(1)
	}
	return node.Generate().Int64()
}

// GenIDString 生成一个全局唯一的字符串 ID
func GenIDString() string {
	if node == nil {
		_ = InitSnowflake(1)
	}
	return node.Generate().String()
}

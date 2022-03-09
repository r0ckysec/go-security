/**
 * @Description
 * @Author r0cky
 * @Date 2021/10/9 18:22
 **/
package job

import (
	cmap "github.com/orcaman/concurrent-map"
)

// 全局map
var (
	Jobmap = cmap.New()
)

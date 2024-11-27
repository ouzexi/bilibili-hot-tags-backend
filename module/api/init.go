package api

import (
	"encoding/json"
	"fmt"
	"gin_template/hub"
	"gin_template/utils"
	"sync"

	"github.com/gin-gonic/gin"
)

type CookieParams struct {
	CookieVal string `json:"cookieVal" binding:"required"`
}

type Mod struct {
}

func (m *Mod) GetModuleInfo() hub.ModuleInfo {
	return hub.ModuleInfo{
		ID:       "internal.ping",
		Instance: m,
	}
}

func (m *Mod) Init() {
	// 初始化过程
	// 在此处可以进行 Module 的初始化配置
	// 如配置读取
}

func (m *Mod) PostInit() {
	// 第二次初始化
	// 再次过程中可以进行跨Module的动作
	// 如通用数据库等等
}

func (m *Mod) Serve(server *hub.Server) {
	// 注册服务函数部分
	server.HttpEngine.GET("/ping", handlePingPong)

	// 发送获取b站标签请求
	server.HttpEngine.GET("/bilibili-tags", handleBilibiliTagsGet)

	// 更新Cookie
	server.HttpEngine.POST("/update-cookie", handleUpdateCookie)
}

func (m *Mod) Start(server *hub.Server) {
	// 此函数会新开携程进行调用
	// ```go
	// 		go exampleModule.Start()
	// ```

	// 可以利用此部分进行后台操作
	// 如http服务器等等
}

func (m *Mod) Stop(server *hub.Server, wg *sync.WaitGroup) {
	// 别忘了解锁
	defer wg.Done()
	// 结束部分
	// 一般调用此函数时，程序接收到 os.Interrupt 信号
	// 即将退出
	// 在此处应该释放相应的资源或者对状态进行保存
}

func handlePingPong(c *gin.Context) {
	c.JSON(200, gin.H{
		"msg":        "pong",
		"User-Agent": c.GetHeader("User-Agent"),
	})
}

func handleBilibiliTagsGet(c *gin.Context) {
	keyword := c.Query("keyword")
	order := c.DefaultQuery("order", "totalrank")

	if keyword == "" {
		c.JSON(500, gin.H{
			"msg": "请输入关键词",
		})
		return
	}

	var responseData utils.ResponseData
	body := utils.FetchVideos(keyword, order)

	if err := json.Unmarshal([]byte(body), &responseData); err != nil {
		c.JSON(500, gin.H{
			"msg": "数据处理出错",
		})
		return
	}
	chartData := utils.TransferRes(responseData)
	fmt.Println(chartData)
	c.JSON(200, gin.H{
		"msg":  "数据获取成功",
		"data": chartData,
	})
}

func handleUpdateCookie(c *gin.Context) {
	var cookieParams CookieParams
	if err := c.ShouldBindJSON(&cookieParams); err != nil {
		c.JSON(500, gin.H{
			"msg": "cookie不能为空！",
		})
		return
	}

	utils.UpdateCookie(cookieParams.CookieVal)
}

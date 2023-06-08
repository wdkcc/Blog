package routers

import (
	"bluebell_backend/controller"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	//gin.SetMode(gin.ReleaseMode)
	//r := gin.New()
	//r.Use(logger.GinLogger(), logger.GinRecovery(true))
	r := gin.Default()
	v1 := r.Group("/api/v1")
	v1.POST("/login", controller.LoginHandler)
	v1.POST("/signup", controller.SignUpHandler)
	v1.GET("/refresh_token", controller.RefreshTokenHandler)

	v1.Use(controller.JWTAuthMiddleware())
	{
		// 获取社区列表
		v1.GET("/community", controller.CommunityHandler)
		// 按ID获取社区信息
		v1.GET("/community/:id", controller.CommunityDetailHandler)
		// 创建帖子
		v1.POST("/post", controller.CreatePostHandler)
		// 按id获取帖子
		v1.GET("/post/:id", controller.PostDetailHandler)
		// 查询能够投票的帖子列表，可选择按时间或者分数排序分页查询(在Redis中查询)
		v1.GET("/post", controller.PostListHandler)
		// Mysql中查询帖子列表
		v1.GET("/post2", controller.PostList2Handler)
		// 投票
		v1.POST("/vote", controller.VoteHandler)
		// 评论
		v1.POST("/comment", controller.CommentHandler)
		// 获取某个帖子的评论信息
		v1.GET("/comment", controller.CommentListHandler)

		v1.GET("/ping", func(c *gin.Context) {
			c.String(http.StatusOK, "pong")
		})

	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404",
		})
	})
	return r
}

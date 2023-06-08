package controller

import (
	"bluebell_backend/dao/redis"
	"bluebell_backend/logic"
	"bluebell_backend/models"
	"fmt"
	"strconv"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// PostHandler 创建帖子
func CreatePostHandler(c *gin.Context) {
	// 1. 得到帖子数据
	var post models.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		ResponseErrorWithMsg(c, CodeInvalidParams, err.Error())
		return
	}
	// 2. 根据当前请求携带的Token获取用户ID,也就是作者的ID
	userID, err := getCurrentUserID(c)
	if err != nil {
		zap.L().Error("GetCurrentUserID() failed", zap.Error(err))
		ResponseError(c, CodeNotLogin)
		return
	}
	post.AuthorId = userID
	// 3. 创建帖子并存储进数据库
	err = logic.CreatePost(&post)
	if err != nil {
		zap.L().Error("logic.CreatePost failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}

// PostListHandler 帖子列表
func PostListHandler(c *gin.Context) {
	order, _ := c.GetQuery("order")
	pageStr, ok := c.GetQuery("page")
	if !ok {
		pageStr = "1"
	}
	pageNum, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil {
		pageNum = 1
	}
	posts := redis.GetPost(order, pageNum)
	fmt.Println(len(posts))
	ResponseSuccess(c, posts)
}

func PostList2Handler(c *gin.Context) {
	data, err := logic.GetPostList2()
	if err != nil {
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)

}

// PostDetailHandler 帖子详情
func PostDetailHandler(c *gin.Context) {
	postId := c.Param("id")

	post, err := logic.GetPost(postId)
	if err != nil {
		zap.L().Error("logic.GetPost(postID) failed", zap.String("postId", postId), zap.Error(err))
	}

	ResponseSuccess(c, post)
}

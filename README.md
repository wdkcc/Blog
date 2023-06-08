# # 主要接口说明

这是一个用来学习和巩固go web项目开发包而开发的博客系统后端程序，支持用户的登陆注册，帖子的创建、分类、评论、投票、按发布时间和投票数量得到热度分数进行排名。

# 快速开始

1. 下载程序代码

   ```git
   git clone https://github.com/2563347014/Blog.git
   ```

   

2. 在Blog/conf/config.yaml中配置mysql和Redis的数据库账号、密码、端口信息。

3. 在Mysql数据库中运行Blog/mysqlcmd/下的sql指令，去创建数据库及数据库表。

4. cd Blog下

5. ```
   go run ./ 运行程序
   ```

# 主要接口说明

## 1、注册登陆

+ 注册：http://127.0.0.1:8081/api/v1/login：POST请求。运行逻辑：接收前端请求数据绑定至结构体->查询用户是否已存在->不存在则使用雪花算法生成用户ID->密码用md5加密后存入数据库->用户数据插入数据库。

  + 注册输入结构体

    ```go
    type RegisterForm struct {
    	UserName        string `json:"username"`
    	Password        string `json:"password"`
    	ConfirmPassword string `json:"confirm_password"`
    }
    ```

    

+ 登陆：http://127.0.0.1:8081/api/v1/login：POST请求。运行逻辑：接收前端输入的用户名和密码->根据用户名在mysql中查询数据记录->验证用户名密码是否正确->正确则使用Jwt生成token并返回给前端（token中包含用户ID字段，因此知道是谁在访问）

## 2、帖子的创建、投票、评论

这些操作需要进行登陆，登陆后携带token才能访问。

+ 创建帖子：http://127.0.0.1:8081/api/v1/post：POST请求。运行逻辑：接收前端传进来的帖子标题、列表、所属社区->token中解析用户ID->雪花算法生成帖子ID->mysql存储详细帖子->redis缓存帖子全部信息并用有序集合zset存储帖子分数、创建时间、所属社区列表，redis具体缓存如下所示：

  + 存储帖子对应用户的投票信息，具体结构是`bluebell:post:voted:postid score userid`

  + 存储帖子投票分数的有序集合，具体结构是`bluebell:post:score score postid`
  
  + 存储帖子时间顺序的有序集合，具体结构是`bluebell:post:time score postid`
  
  + 存储帖子详细信息的集合，具体结构是`bluebell:post:postid 帖子信息`。
  
    ```go
    // 帖子的结构体
    type Post struct {
    	PostID      uint64    `json:"post_id" db:"post_id"`
    	Title       string    `json:"title" db:"title"`
    	Content     string    `json:"content" db:"content"`
    	AuthorId    uint64    `json:"author_id" db:"author_id"`
    	CommunityID int64     `json:"community_id" db:"community_id"`
    	Status      int32     `json:"status" db:"status"`
    	CreateTime  time.Time `json:"-" db:"create_time"`
    }
    ```

+ 投票：http://127.0.0.1:8081/api/v1/vote：POST请求。运行逻辑：获取用户投票帖子ID、投票方向、用户ID->判断帖子是否过期，过期将不能再投票->获取用户之前对帖子投票的信息->更新用户对该帖子的投票并更新帖子分数。（帖子分数按照帖子发布时间，当前累积点赞数量进行计算）

+ 评论：http://127.0.0.1:8081/api/v1/comment：POST请求。运行逻辑：获取用户评论信息、评论的帖子ID->雪花算法生成评论ID->根据Token获取发表评论的用户ID->存入数据库。

# 参考

[Go Web开发进阶实战（gin框架）](https://study.163.com/course/courseMain.htm?courseId=1210171207)


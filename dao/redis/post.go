package redis

import (
	"math"
	"time"

	"github.com/go-redis/redis"
)

const (
	OneWeekInSeconds         = 7 * 24 * 3600
	VoteScore        float64 = 432
	PostPerAge               = 20
)

/*
投票算法：http://www.ruanyifeng.com/blog/2012/03/ranking_algorithm_reddit.html



*/

/*
	PostVote 为帖子投票

投票分为四种情况：1.投赞成票 2.投反对票 3.取消投票 4.反转投票

记录文章参与投票的人
更新文章分数：赞成票要加分；反对票减分

v=1时，有两种情况

	1.之前没投过票，现在要投赞成票
	2.之前投过反对票，现在要改为赞成票

v=0时，有两种情况

	1.之前投过赞成票，现在要取消
	2.之前投过反对票，现在要取消

v=-1时，有两种情况

	1.之前没投过票，现在要投反对票
	2.之前投过赞成票，现在要改为反对票
*/
func PostVote(postID, userID string, v float64) (err error) {
	// 1. 取帖子发布时间
	postTime := client.ZScore(KeyPostTimeZSet, postID).Val()
	// 2. 帖子已过期，不能再投票
	if float64(time.Now().Unix())-postTime > OneWeekInSeconds {
		return ErrorVoteTimeExpire
	}
	// 3. 获取之前帖子的分数
	key := KeyPostVotedZSetPrefix + postID
	ov := client.ZScore(key, userID).Val() // 获取当前分数
	diffAbs := math.Abs(ov - v)
	pipeline := client.TxPipeline()
	// 4. 记录当前投票
	pipeline.ZAdd(key, redis.Z{ // 记录已投票
		Score:  v,
		Member: userID,
	})
	// 5. 更新分数
	pipeline.ZIncrBy(KeyPostScoreZSet, VoteScore*diffAbs*v, postID) // 更新分数

	//switch math.Abs(ov) - math.Abs(v) {
	//case 1:
	//	// 取消投票 ov=1/-1 v=0
	//	// 投票数-1
	//	pipeline.HIncrBy(KeyPostInfoHashPrefix+postID, "votes", -1)
	//case 0:
	//	// 反转投票 ov=-1/1 v=1/-1
	//	// 投票数不用更新
	//case -1:
	//	// 新增投票 ov=0 v=1/-1
	//	// 投票数+1
	//	pipeline.HIncrBy(KeyPostInfoHashPrefix+postID, "votes", 1)
	//default:
	//	// 已经投过票了
	//	return ErrorVoted
	//}
	_, err = pipeline.Exec()
	return
}

// CreatePost 使用redis的有序集合存储帖子的分数信息和时间信息，并将帖子id存储在对应的社区列表中
func CreatePost(postID, userID, title, summary, communityName string) (err error) {
	now := float64(time.Now().Unix())
	votedKey := KeyPostVotedZSetPrefix + postID
	communityKey := KeyCommunityPostSetPrefix + communityName
	postInfo := map[string]interface{}{
		"title":    title,
		"summary":  summary,
		"post:id":  postID,
		"user:id":  userID,
		"time":     now,
		"votes":    1,
		"comments": 0,
	}

	// 事务操作
	pipeline := client.TxPipeline()
	// 1. 存储帖子中用户的投票信息
	pipeline.ZAdd(votedKey, redis.Z{ // 作者默认投赞成票
		Score:  1,
		Member: userID,
	})
	pipeline.Expire(votedKey, time.Second*OneWeekInSeconds) // 设置过期一周时间
	// 2. 存储帖子详细信息
	pipeline.HMSet(KeyPostInfoHashPrefix+postID, postInfo)
	// 3. 添加投票分数数据
	pipeline.ZAdd(KeyPostScoreZSet, redis.Z{
		Score:  now + VoteScore,
		Member: postID,
	})
	// 4. 添加时间数据
	pipeline.ZAdd(KeyPostTimeZSet, redis.Z{ // 添加到时间的ZSet
		Score:  now,
		Member: postID,
	})
	// 5. 添加帖子ID到对应社区下
	pipeline.SAdd(communityKey, postID)
	_, err = pipeline.Exec()
	return
}

// GetPost 从key中分页取出帖子
func GetPost(order string, page int64) []map[string]string {
	// 默认按帖子分数排序
	key := KeyPostScoreZSet
	// 按创建时间排序
	if order == "time" {
		key = KeyPostTimeZSet
	}
	start := (page - 1) * PostPerAge
	end := start + PostPerAge - 1
	// 查询得到帖子id
	ids := client.ZRevRange(key, start, end).Val()
	postList := make([]map[string]string, 0, len(ids))
	for _, id := range ids {
		postData := client.HGetAll(KeyPostInfoHashPrefix + id).Val()
		postData["id"] = id
		postList = append(postList, postData)
	}
	return postList
}

// GetCommunityPost 分社区根据发帖时间或者分数取出分页的帖子
func GetCommunityPost(communityName, orderKey string, page int64) []map[string]string {
	key := orderKey + communityName // 创建缓存键

	if client.Exists(key).Val() < 1 {
		client.ZInterStore(key, redis.ZStore{
			Aggregate: "MAX",
		}, KeyCommunityPostSetPrefix+communityName, orderKey)
		client.Expire(key, 60*time.Second)
	}
	return GetPost(key, page)
}

// Reddit Hot rank algorithms
// from https://github.com/reddit-archive/reddit/blob/master/r2/r2/lib/db/_sorts.pyx
func Hot(ups, downs int, date time.Time) float64 {
	s := float64(ups - downs)
	order := math.Log10(math.Max(math.Abs(s), 1))
	var sign float64
	if s > 0 {
		sign = 1
	} else if s == 0 {
		sign = 0
	} else {
		sign = -1
	}
	seconds := float64(date.Second() - 1577808000)
	return math.Round(sign*order + seconds/43200)
}

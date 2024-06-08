package redis

import (
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

const (
	oneWeekInseconds = 7 * 24 * 3600
	scorePerVote     = 443
)

var (
	ErrVoteTimeExpire = errors.New("投票时间已过")
	ErrVoteRepeated   = errors.New("不允许重复投票")
)

// 用户投票的相关算法： 阮一峰的博客

//本项目使用简化版的投票分数

/* 投票的几种情况:
direction=1时，有两种情况：
	1.之前没有投过票，现在投赞成票
	2.之前投反对票，现在改投赞成票
direction=0时，有两种情况：
	1.之前投过赞成票，现在要取消投票
	2.之前投过反对票，现在要取消投票
direction=-1时，有两种情况:
	1.之前没有投过票，现在投反对票
	2.之前投过赞成票，现在改投反对票

投票的限制：
每个帖子自发表之日起一个星期之内允许用户投票，超过一个星期求不允许再投票了。
	1.到期之后将redis中保存的赞成票数及反对票数存储到mysql表中
	2.到期之后删除 KeyPostVotedZSetPF
*/

func CreatePost(postID, communityID int64) error {

	//创建事务
	pipeline := rdb.TxPipeline()
	//帖子时间
	pipeline.ZAdd(getRedisKey(KeyPostTimeZSet), redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: postID,
	})

	//帖子分数
	pipeline.ZAdd(getRedisKey(KeyPostScoreZSet), redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: postID,
	})

	// 把帖子id加到社区的set
	cKey := getRedisKey(KeyCommunitySetPF + strconv.Itoa(int(communityID)))
	pipeline.SAdd(cKey, postID)

	_, err := pipeline.Exec()
	return err
}

func VoteForPost(userID, postID string, value float64) error {
	// 1. 判断投票限制
	// 去redis取帖子发布时间
	postTime := rdb.ZScore(getRedisKey(KeyPostTimeZSet), postID).Val()
	if float64(time.Now().Unix())-postTime > oneWeekInseconds {
		return ErrVoteTimeExpire
	}
	// 2. 更新帖子的分数
	// 先查看当前用户给当前帖子投票的记录
	ov := rdb.ZScore(getRedisKey(KeyPostVotedZSetPF+postID), postID).Val()
	if value == ov {
		return ErrVoteRepeated //防止重复投票
	}
	var op float64
	if value > ov {
		op = 1
	} else {
		op = -1
	}
	diff := math.Abs(ov - value) // 计算两次投票的差值
	pipeline := rdb.TxPipeline()
	pipeline.ZIncrBy(getRedisKey(KeyPostScoreZSet), op*diff*scorePerVote, postID)

	// 3. 记录用户为该帖子投票的数据
	if value == 0 {
		rdb.ZRem(getRedisKey(KeyPostVotedZSetPF+postID), postID).Result()
	} else {
		rdb.ZAdd(getRedisKey(KeyPostVotedZSetPF+postID), redis.Z{
			Score:  value,
			Member: userID,
		}).Result()
	}
	_, err := pipeline.Exec()
	return err
}

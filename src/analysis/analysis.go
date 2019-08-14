package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mgutz/str"
	"github.com/sirupsen/logrus"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	HEADLE_DIG = " /dig?"
 	HEADLE_HTTP = " HTTP/"
 	HEADLE_JSP = ".jsp"
 	HEADLE_GEEKTECH = "/GeekTech/"
 	//可扩展的过滤信息点 HEADLE_XXX
 	)



//与js客户端收集的json数据一致
type digData struct{
	time string
	ip   string
	currentUrl string
	referUrl   string
	userAgent  string
}

//用于在pvChannel或者uvChannel中传输数据用的struct
type urlData struct{
	data digData  //digData就是URL中的参数链的数据，以digData结构体存储
	uid  string   //用户唯一标识，用于UV统计，去重
	unode urlNode
}

//传输到storageChannel中的具体数据的struct
//想在redis中存储怎么样的数据，就在urlNode这里定义
type urlNode struct{
	unType  string  //urlNode的类型：如详情页、首页、还是列表页等
	unRid   string  //resource的id，一般为页面的名字，如果只有数字就用int，如果有字母就用string
	unUrl   string  //当前访问的页面的URL
	unTime  string  //当前访问这个URL的时间
}

//用于在storageChannel通道传输的结构体struct
type storageBlock struct{
	counterType  string  //统计类型，如PV或者是UV、或是其他统计类型
	storageModel string  //存储的数据格式，即redis的命令
	unode 		 urlNode //需要被存储器存储的具体数据
}

type cmdParams struct{
	logFilePath string  //要消费的日志文件路径
	routineNum  int    //自定义的并发度
}

//日志的级别：debug、info、warn、error、fatal、panic 【由低到高】
//级别越低，打印的日志信息越详细越多，debug模式一般用于开发环境
//error就是发生错误还能运行，fatal和panic是发生错误不能继续运行了吧

var log = logrus.New()  //实例化一个全局的logrus对象
//var redisCli redis.Client  //声明一个结构体
func init() {
	log.Out = os.Stdout  //log的输出，就适用os包中的标准输出就行了
	log.SetLevel( logrus.DebugLevel )
	//redisCli, err := redis.Dial("tcp", "localhost:6379")
	//if err != nil{
	//	log.Fatal("连接不上redis处理引擎... ==> " , err.Error())
	//}
	//defer redisCli.Close()

	//多个goroutine一起使用一个redis-client实例的时候，用连接池效果更好
}


func main(){
	//读取命令行参数
	//args: 1）参数名   2）参数的默认值   3）参数的含义说明
	logFilePath := *flag.String("logFilePath", "../logs/dig.log", "log file path")
	routineNum := *flag.Int("routineNum", 5, "routine的并发度")
	targetLog := *flag.String("targetLog", "../logs/runtime.log", "打印日志的所在位置")
	flag.Parse()

	params := cmdParams{logFilePath:logFilePath, routineNum:routineNum}

	//打印日志 - 日志等级
	logfd, err := os.OpenFile(targetLog, os.O_CREATE | os.O_WRONLY, 0644)
	if err != nil {
		log.Errorf("打开文件:%s 错误 错误信息："+err.Error(), targetLog)
	}
	log.Out = logfd  //使用log文件句柄作为Out输出
	defer logfd.Close()  //关闭文件句柄

	log.Infoln("Exec start. ")  //向runtime.log文件中写入数据
	log.Infof("cmdParams: logFilePath=%s , routineNum=%d ", params.logFilePath, params.routineNum)


	//初始化一些buffered channel来在各个goroutine间传递数据
	logChannel := make(chan string, 3*params.routineNum)  //日志消费者与日志解析组之间的channel，传输数据量比较大
	pvChannel := make(chan urlData, params.routineNum)  //日志解析组与pv统计器之间的channel
	uvChannel := make(chan urlData, params.routineNum)  //日志解析组与uv统计器之间的channel
	storageChannel := make(chan storageBlock, params.routineNum)  //pv、uv统计器之间的channel


	//redis连接池，用于多个goroutine一起并发使用redis-client的情况下，效果会更好
	//使用redis pool: tcp ,ip地址，并发度
	//得到redis pool的实例
	redisPool, err := pool.New("tcp", "127.0.0.1:6379", 1)
	if err != nil{ //连接时发生错误
		log.Fatalln("pool.New()时发生错误 ==> ", err.Error())
		panic(err)
	}else{ //成功打开连接池后，若没有流量访问redis的情况下，一段时间内，连接池会自动关闭
	//所以需要用一个goroutine专门3秒去ping一下redisPool，以保持没有流量下，在并发场景下的持续连接
		go func(){
			for{
				redisPool.Cmd("PING")
				time.Sleep(3*time.Second)
			}
		}()
	}


	//创建一个日志消费者
	go  ReadLogFileLineByLine(params, logChannel)

	//创建一组日志解析者
	for i:=0; i <= params.routineNum; i++ {
		go LogConsumer(logChannel, pvChannel, uvChannel)
	}

	//PV UV统计器
	go pvCounter( pvChannel, storageChannel )
	go uvCounter( uvChannel, storageChannel, redisPool )
	//可扩展统计器

	//存储器
	go DataStorage( storageChannel, redisPool )

	time.Sleep( 1000 * time.Second ) //开发期间使用sleep，防止goroutine前提退出，就是要执行1000秒后,sleep才失去效果

}

//日志消费者，即去一行一行读取日志的goroutine
func ReadLogFileLineByLine(params cmdParams, logChannel chan string) error {

	//log.Infoln("====> this is readlogFile goroutine ===>")

	fd, err := os.OpenFile(params.logFilePath, os.O_RDONLY, 0644)
	if err != nil{
		log.Errorf("func:ReadLogFileLineByLine can't open the file: %s", params.logFilePath)
	}
	defer fd.Close()

	count := 0 //计算器，用于记录已经读取到第几行了
	bufferRead := bufio.NewReader(fd)  //New一个针对fd文件句柄的Reader，专门读取fd指向的文件内容
	for {  //有日志就读，所以是死循环，要让CPU进行调度
		line, err := bufferRead.ReadString( byte('\n') )
		if err != nil{
			if err == io.EOF {  //说明文件已经读取完毕了，或者说是log文件里面没有内容了
				time.Sleep( 3 * time.Second ) //让此程序睡上3秒，在用户态空间让出运行权，让其他goroutine继续并发执行
				log.Infof("func:ReadLogFileLineByLine wait, line=%d", count)
			}else{ //发生其他错误
				log.Warningln("func:ReadLogFileLineByLine read dig.log error ==> " + err.Error())
			}
		}
		count ++

		//fmt.Println(line) 数据到这了


		//logChannel的buffer只有15个string的slice
		//所以当logChannel中满15个string的时候，写入的goroutine就会停止，即被阻塞
		logChannel <- line

		if count % ( 1000 * params.routineNum ) == 0 { //每读1000行，向runtime.log中写入提示消息
			log.Infof("func:ReadLogFileLineByLine line: %d", count)
		}
	}

	return nil
}

//日志解析者组，即去解析日志的一组goroutine
func LogConsumer(logChannel chan string, pvChannel, uvChannel chan urlData) error {

	//log.Infoln("====> this is logConsumer goroutine ===>")

	for logStr := range logChannel {  //逐行消费logChannel中的string

		//fmt.Println("LogChannel -> ", logStr) this

		//切割日志，过滤出打点上报的信息，即digData结构体数据
		digData := cutLogFetchData( logStr )

		//uid: 使用md5(referUrl + userAgent) => hash值,当hash值一样的情况下，就认为是同一个用户
		hasher := md5.New()  //返回一个Hash接口

		//为什么要这样生成UID呢？不应该是ip地址MD5一下吗？因为ip地址是不变的，所以需要这样做
		hasher.Write([]byte(digData.referUrl + digData.userAgent + strconv.FormatInt(time.Now().UnixNano(), 10))) //理论上可以输入无限大的数据
		uid := hex.EncodeToString(hasher.Sum(nil))  //把[]byte流转换为Hex十六进制的HASH值

		//fmt.Println("currentUrl=",digData.currentUrl)
		//fmt.Println("time=",digData.time)

		//拼出urlData用于pvChannel和uvChannel的传输
		uData := urlData{digData, uid, formatUrl(digData.currentUrl, digData.time)}


		//fmt.Println("unode: ", uData.unode)

		uvChannel <- uData
		pvChannel <- uData

		// XXXChannel <- uData , 在这可扩展统计器Channel
	}

	return nil
}

//用于生成urlNode结构体
func formatUrl( url, t string) urlNode {
	//从不同类型页面中，页面数量最大的着手


	pos1 := str.IndexOf(url, HEADLE_GEEKTECH, 0)
	if pos1 == -1 { //匹配失败
		return urlNode{}
	}

	//说明匹配到了，提取出所有页面的pageId，包括主页的index.jsp的ID，即index
	pos1 = pos1 + len(HEADLE_GEEKTECH)
	pos2 := str.IndexOf(url, HEADLE_JSP, pos1)
	pageId := str.Substr(url, pos1, pos2-pos1) //从pos1位置开始，截取pos2-pos1长度的字符
	//pageId, _ := strconv.Atoi(idStr)    //页面id或者是页面的name

	//fmt.Println("pageId=", pageId)


	//封装一个urlNode
	return urlNode{
		unType: "allPages",   //页面的类型，我这里的网站类型太少，就用这个表示吧
		unRid:  pageId+".jsp",   //页面的名字，或页面id
		unUrl:  url,           //当前页面的URL
		unTime: t,             //访问当前页面的时间
	}
}

//辅助函数，用于从一长串的一行log中提出URL，并存储在k-v结构中，最后封装到digData结构体中
//即提取出URL网址字符串，然后进行URL的decoding解码操作，提出k-v的数据
func cutLogFetchData( logStr string ) digData {
	logStr = strings.TrimSpace(logStr)  //去掉logStr前后两端的空格

	pos1 := str.IndexOf(logStr, HEADLE_DIG, 0) //从0开始搜索到HEADLE_DIG字符串为止，返回的是HEADLE_DIG字符串的第一个字符的index
	if pos1 == -1 { //没有匹配到
		return digData{}
	}

	pos1 = pos1 + len(HEADLE_DIG) //定位到ip的i字符位置
	pos2 := str.IndexOf(logStr, HEADLE_HTTP, pos1)
	argsData := str.Substr(logStr, pos1, pos2-pos1)  //最后一个参数是要取多少个字符，取pos2-pos1这个长度的字符串

	//解析一个URL结构，把里面的参数都解析出来，即都解码decode一遍
	//想拼成&ip=xxx&time=xxx，就使用Value{}.encode编码出来
	urlObj, err := url.Parse("http://127.0.0.1:8000/GeekTech/index.jsp?" + argsData)  //解析URL结构，并返回一个解析后的URL对象
	if err != nil {
		return digData{}
	}
	dataVal := urlObj.Query()  //把urlObj转换为Value结构体，Value结构体是存储k-v的结构体


	//fmt.Println(" ==> ", dataVal.Get("referUrl"))
	//fmt.Println(dd)

	return digData{
		ip:dataVal.Get("ip"),
		time:dataVal.Get("time"),
		currentUrl:dataVal.Get("url"),
		referUrl:dataVal.Get("referUrl"),
		userAgent:dataVal.Get("userAgent"),
	}
}

//PV统计器
func pvCounter( pvChannel chan urlData, storageChannel chan storageBlock ) {

	//log.Infoln("====> this is pvCounter goroutine ===>")

	for dataItem := range pvChannel{ //逐个从队列中取出
		storageItem := storageBlock{"pv", "ZINCRBY", dataItem.unode}

		//fmt.Println(storageItem)

		storageChannel <- storageItem  //写入storageChannel
	}
}

//UV统计器
func uvCounter( uvChannel chan urlData, storageChannel chan storageBlock, redisPool *pool.Pool ) {
	//业内统计日活跃独立用户数的方法，使用Hyperloglog结构
	//Hyperloglog内的元素是无重复元素的
	//HyperLogLog API：pfadd key elem1,elem2 ...  |  pfcount key  | pfmerge newKey key1 key2 ，去重操作
	//HyperLogLog缺点：用极小的空间去统计周期活跃用户数 ，官方给出的错误率是 0.81%，而且不能取出单条数据
	//具统计：百万用户量级一天只需15KB，一个月只需30*15=450KB,一年大约也是365*15KB=5MB左右

	//log.Infoln("====> this is uvCounter goroutine ===>")

	//UV按天去重
	//count := 0
	//insertCount := 0
	for dataItem := range uvChannel{

		//fmt.Println(dataItem)

		//hpll = hyperloglog
		HyperLogLogKey := "uv_hyperloglog_day_" + getTime("day")  //hyperloglog结构的key

		//zincrby key 1 uid 或者 pfadd key uid
		res, err := redisPool.Cmd("pfadd", HyperLogLogKey, dataItem.uid).Int()  //调用一次cmd，只算一次插入
		//insertCount ++
		//fmt.Println("insertCount = ", insertCount)
		if err != nil{
			log.Warningln("redisPool.Cmd(hyperloglog...) 发生错误 ==> ", err.Error())
		}

		if res <= 0{ //插入不成功继续循环
			//count ++
			//fmt.Println(count)
			continue
		}

		//fmt.Println(count)


		//fmt.Println("uvCouner ==> 结构体：" , dataItem.unode.unType , " == >" ,dataItem.unode.unUrl)

		//插入hyperloglog结构成功，再生成一个StorageBlock写入StorageChannel
		sItem := storageBlock{"uv", "ZINCRBY", dataItem.unode}

		storageChannel <- sItem

	}

}

//当前时间生成器
//日志中记录的网页访问的时间和时间类型
//业内固定写法
func getTime(timeType string) string {
	var shortForm string
	switch timeType {
	case "day":
		shortForm = "2006-01-02"
		break
	case "hour":
		shortForm = "2006-01-02 15"
		break
	case "minute":
		shortForm = "2006-01-02 15:04"
		break
	case "second":  //一般用不到second级别的
		shortForm = "2006-01-02 15:04:05"
		break
	}

	t, _ := time.Parse(shortForm, time.Now().Format(shortForm))
	return strconv.FormatInt(t.Unix(), 10) //把64位的Unix时间戳变为十进制的string
}

func DataStorage( storageChannel chan storageBlock, redisPool *pool.Pool) {

	//log.Infoln("====> this is DataStorage goroutine ===>")

	//zset有序集合：key -> score:elemId, score可以重复, elemId域不能重复
	//API: 添加zadd key score1 elem1, score2 elem2, score3 elem3 =>O(logN)
	//删除元素：zrem key elemId   获取分数：zscore key elemId  =>O(1)
	//给元素增加和删除指定分值：zincrby key scoreVal elemId => zincrby key 1 tom, zincrby key -1 tom =>O(1)
	//返回元素总个数：zcard key  => O(1)
	//获取元素的排名：zrank key elemId , 从小到大的一个排序顺序
	//获取所有元素的排名：zrange key 0 -1 withscores  => O((logN)+M)


	// 1000万行数据 * N(统计器类型) * M(加洋葱皮的过程)
	//逐层添加，洋葱皮 => 访问了/moive/list/123.html ,就要统计/movie、/movie/list、/movie/list/123.html
	//存储的时候要区分纬度：天 - 小时 - 分钟
	//存储的时候要区别层级：首页 - 大分类 - 小分类A - 小小分类A - .... - 终极页面或者最终目标
	//存储模型： redis sortedSet ,有序集合
	//count, inserCount := 0, 0
	for block := range storageChannel {

		//fmt.Println(block)

		//逐个从storageChannel中消费,因为block中的counterType不同
		prefix := block.counterType + "_" //前缀uv_ 或者pv_ 或者其他的XX_等

		//fmt.Println("结构体：" , block.unode.unType , " == >" ,block.unode.unUrl, " ==> ", block.unode.unTime)

		//逐层添加洋葱皮
		setKeys := []string{  //以为分别为，有序集合是keys
			//时间维度
			prefix + "day_" + getTime("day"),  //一天之内pv的all pages的访问量，一天之内all pages的访问人数
			prefix + "hour_" + getTime("hour"), //一个小时之内pv或者uv的所有页面的访问量
			prefix + "minute_" + getTime("minute"),  //一分钟之内pv或者uv的所有页面的访问量

			//页面层级维度
			prefix + block.unode.unType + "_day_" + getTime("day"), //pv一天之内某个类中的所有页面的访问量
			prefix + block.unode.unType + "_hour_" + getTime("hour"),
			prefix + block.unode.unType + "_minute_" + getTime("minute"),
		}

		//sortedSet中的member域, 而不是score域
		pageId := block.unode.unRid  //对应网页的网页名或者网页id号，不可有重复元素

		//循环向SortedSet中插入key-value数据

		for _, key := range setKeys{
			//pv的含义：一天之内某个类的某个页面被访问了多少次，即浏览量是多少
			//uv的含义： 一天之内某个类的某个页面被多少个独立用户访问了，即用户量是多少
			res, err := redisPool.Cmd(block.storageModel, key, 1, pageId).Int()

			if res <= 0 || err != nil{
				log.Errorln("DataStorage redis error! ==> ", err.Error() , " ==> 定位：", block.storageModel, key, pageId )
			}
		}

		//1）如查询某个页面，每天的访问量是多少
		//zscore key pageId  => 返回的数，就是访问量，如： zscore pv_day_123891 index.jsp  => 1000

		//2）如查询某个页面，每天的访问人数是多少
		//zscore key pageId  => 返回的数，就是访问人数，如： zscore uv_day_123891 index.jsp  => 1000

		//pfadd key e1 e2 e3, Hyperloglog用于统计独立用户数  =>

	}
}

























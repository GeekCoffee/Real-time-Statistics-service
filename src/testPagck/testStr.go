package main

import (
	"fmt"
	"strconv"
	"time"
)

const HEADLE_DIG = " /dig?"

//var log = logrus.New()  //实例化一个全局的logrus对象
//func init(){
//	log.SetLevel( logrus.DebugLevel )
//	log.Out = os.Stdout    //得到OS的标准输出
//}

func main(){
	//logStr := "172.20.10.4 - - [23/Jul/2019:12:28:48 +0800] \"GET /dig?ip=172.20.10.14&refuerUrl=http%3A%2F%2F192.168.200.33%2FGeekTech%2FWeb.jsp&time=2019-07-23+13%3A09%3A24&url=http%3A%2F%2F192.168.200.33%2FGeekTech%2FWeb.jsp&userAgent=Mozilla%2F5.0+%28Linux%3B+Android+4.1.2%3B+Nexus+7+Build%2FJZ054K%29+AppleWebKit%2F535.19+%28KHTML%2C+like+Gecko%29+Chrome%2F18.0.1025.166+Safari%2F535.19  HTTP/1.1\" 200 43 \"http://192.168.200.33/GeekTech/Web.jsp\" Mozilla/5.0 (Linux; Android 4.1.2; Nexus 7 Build/JZ054K) AppleWebKit/535.19 (KHTML, like Gecko) Chrome/18.0.1025.166 Safari/535.19 \"- \" "
	//logStr = strings.TrimSpace(logStr)  //去掉logStr前后两端的空格
	//pos1 := str.IndexOf(logStr, HEADLE_DIG, 0) //从0开始搜索到HEADLE_DIG字符串为止，返回的是HEADLE_DIG字符串的第一个字符的index
	//if pos1 == -1 { //没有匹配到
	//	//return {}
	//}
	//
	//pos1 =pos1 + len(HEADLE_DIG)
	//pos2 := str.IndexOf(logStr, " HTTP/", pos1)
	//argsData := str.Substr(logStr, pos1, pos2-pos1)  //最后一个参数是要取多少个字符，取pos2-pos1这个长度的字符串
	//
	////解析一个URL结构，把里面的参数都解析出来，即都解码decode一遍
	//urlObj, _ := url.Parse("http://127.0.0.1:8000/?" + argsData)  //解析URL结构，并返回一个解析后的URL对象
	//dataVal := urlObj.Query()  //把urlObj转换为Value结构体，Value结构体是存储k-v的结构体
	//ip := dataVal.Get("ip")
	//time := dataVal.Get("time")
	//referUrl := dataVal.Get("referUrl")
	//currentUrl := dataVal.Get("url")
	//userAgent := dataVal.Get("userAgent")
	//
	//fmt.Println(ip, " ", time, " ", referUrl, " ")
	//fmt.Println(currentUrl, " ", userAgent, " ")

	//data := []byte("hello")
	//hashRes := md5.Sum(data) //这个[16]byte不能转换为string
	////hashStr := hex.EncodeToString([]byte(hashRes))
	//fmt.Printf("%x\n", hashRes)
	//
	//hasher := md5.New() //得到一个hash接口
	//hasher.Write(data)
	//hashRes2 := hasher.Sum(nil)
	//hashStr := hex.EncodeToString(hashRes2) //这个[]byte可以转换为string
	//fmt.Printf("%x\n", hashRes2)
	//fmt.Println(hashStr)

	const shortForm = "2006-01-02 15"  //固定写法
	t, _ := time.Parse(shortForm, time.Now().Format(shortForm))
	fmt.Println(t)
	fmt.Println(strconv.FormatInt(t.Unix(),10))

	//testLog := *flag.String("testLog", "../../logs/testLog.log", "打印日志的所在位置")
	//


	//logfd, err := os.OpenFile("hello.log",  os.O_CREATE | os.O_WRONLY, 0666)
	//if err != nil{
	//	//	fmt.Println("打开文件失败！" + err.Error())
	//	//}
	//	//log.Out = logfd
	//	//defer logfd.Close() //close file handlle
	//	//
	//	//
	//log.Infoln("Exec start. ")
	//log.Infof("只要调用了log.Info()就可以把这行信息写入到或者叫记录到指定的日志中")
	//log.Infoln("hello goroutine!")


	//redisPool, err := pool.New("tcp", "127.0.0.1:6379",2)
	//if err != nil{ //连接时发生错误
	//	fmt.Println("pool.New()时发生错误 ==> ", err.Error())
	//	panic(err)
	//}else{ //成功打开连接池后，若没有流量访问redis的情况下，一段时间内，连接池会自动关闭
	//	//所以需要用一个goroutine专门3秒去ping一下redisPool，以保持没有流量下，在并发场景下的持续连接
	//	//go func(){
	//	//	for{
	//	//		redisPool.Cmd("PING")
	//	//		time.Sleep(3*time.Second)
	//	//	}
	//	//}()
	//}
	//res,err := redisPool.Cmd("SET", "golang", "goroutine").Int()
	//if err != nil{
	//	fmt.Println( " ===>", err)
	//}
	//fmt.Println("res = ", res)


}

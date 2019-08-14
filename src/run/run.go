package main

//run.go文件的功能是实现模拟用户访问网站的行为，写压力测试代码

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"time"
)

//资源结构体
type resource struct {
	url string
	target string
	start int  //start和end预留着，以备URL中有数字
	end int
}

//把实体资源数据存入list中并返回
func ruleResource() []resource {
	var resList []resource
	//ip := "172.20.10.14:8080" 手机热点
	ip := "192.168.200.33"  //WiFi热点

	//index.jsp
	r1 := resource{
		url : "http://" + ip + "/GeekTech/{$pageName}.jsp",
		target:"{$pageName}",
		start:0,
		end:0,
	}

	resList = append(resList, r1)
	return resList

	//可以增加的resource
	//r2 := resource{
	//	url : "http://" + ip + "/GeekTech/{$pageName}.jsp",
	//	target:"{$pageName}",
	//	start:0,
	//	end:0,
	//}
}

func buildUrl( res []resource ) []string {
	var urlList []string  //记录URL的list
	var pageName = []string{"index", "aboutUs","AR","BigData","Unity3D","VR","Web","weixin","yidong"}


	//遍历resource资源结构体数组
	for _, resItem := range res {
		if len(resItem.target) != 0 { //替换URL工作
			for _, name := range pageName {
				//源string、old string、new string, -1为不重复替换-只替换1次
				urlStr := strings.Replace(resItem.url, resItem.target, name, -1)
				urlList = append(urlList, urlStr)
			}
		}
	}
	return urlList
}

//格式化输出一行日志内容
func makeLog(currentUlr, referUrl, userAgent, ip string) string {
	u := url.Values{}  //构造多个k-v组成的参数list
	u.Set("time", "2019-07-23 13:09:24")  //用户点击网页的时间
	u.Set("ip", ip)  //用户ip
	u.Set("url", currentUlr) //用户所在的当前页面的URL
	u.Set("referUrl", referUrl)  //用户所在当前页面的URL的前一个URL
	u.Set("userAgent", userAgent)  //客户端浏览器信息
	paramsStr := u.Encode()  //拼起来做一次URL的encode，返回如：time=xxx&ip=172.0.0.1&url=XXX&referUrl=XXX&userAgent=XXX

	//time, ip, url ,referUrl, ua
	logTemplate := "172.20.10.4 - - [23/Jul/2019:12:28:48 +0800] \"GET /dig?{$paramsStr}  HTTP/1.1\" 200 43 \"{$currentUrl}\" {$userAgent} \"- \" "
	log := strings.Replace(logTemplate, "{$paramsStr}", paramsStr, -1)
	log = strings.Replace(log, "{$userAgent}", userAgent, -1)
	log = strings.Replace(log, "{$currentUrl}", currentUlr, -1)

	fmt.Println("log ==> ", log)
	return log //返回的是拼凑起来的一行日志
}

//使用时间作为随机种子，在不同时间下随机出来的结果是不一样的
//不设置随机种子的情况下，就会出现伪随机数
func randInt(min, max int) int {
	if min >= max {  //检测传入是否为负数
		return max
	}

	r := rand.New(rand.NewSource( time.Now().UnixNano() ))  //传入Unix时间戳为随机种子

	//rand.Intn(args)的随机值在[0,n)区间
	return r.Intn(max-min+1) + min //这样写更具通用性，增加偏移量可以让随机值在[min,max]这个区间
}

func main(){

	//flag.Int()用于接受命令行传来的参数
	//参数名、参数的值、说明描述
	total := flag.Int("total", 100, "从命令行获取参数，动态生成网页访问量的程序...")
	filePath := flag.String("filePath", "../logs/dig.log", "log file path")
	flag.Parse()  //从命令行中解析参数，把数据放入相应的变量中，底层调用了os.Args[1:]

	//fmt.Println(*total, *filePath)

	//step1：需要先取构造网站的所有真实URL的集合,可以使用slice结构存储
	resList := ruleResource()  //资源结构体的list
	urlList := buildUrl(resList) //url的list

	//fmt.Println(urlList)

	//step2：源自以上构造出来的list，生成total行日志
	logStr := ""
	for i := 0; i <= *total; i ++ {
		currentUrl := urlList[ randInt(0, len(urlList)-1) ]
		referUrl := urlList[ randInt(0, len(urlList)-1) ]

		//fmt.Println("referUrl => ", referUrl)

		userAgent := userAgentList[ randInt(0, len(userAgentList)-1) ]

		ip := "172.20.10.4" //测试用例，使用一个固定ip就行了
		logStr = logStr + makeLog(currentUrl, referUrl, userAgent, ip) + "\n"

		//WriteFile函数的功能：把数据写入指定文件中，在这里是写入dig.log中，覆盖写入
		//perm是文件写入模式
		//ioutil.WriteFile(*filePath, []byte(logStr), 0644)
	}

	//一次性追加写入文件中， O_RDWR是指read-write模式 | O_CREATE是文件不存在创建一个 | os.O_APPEND追加到文件末尾写入
	fd, err := os.OpenFile(*filePath, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0644) //文件权限
	if err != nil {
		log.Fatal(err.Error())  //打印错误信息
	}

	fd.Write([]byte(logStr))
	fd.Close()

	fmt.Println("done. \n")

	//fmt.Println("url max index: ", len(urlList)-1)
	//for i := 0; i < 100; i++{
	//	fmt.Println(randInt(0, len(urlList)-1))
	//}

}


//常用的userAgent收集
var userAgentList = []string{

	//Android平台原生浏览器
	"Mozilla/5.0 (Linux; Android 4.1.1; Nexus 7 Build/JRO03D) AppleWebKit/535.19 (KHTML, like Gecko) Chrome/18.0.1025.166  Safari/535.19",
	"Mozilla/5.0 (Linux; U; Android 4.0.4; en-gb; GT-I9300 Build/IMM76D) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/534.30",
	"Mozilla/5.0 (Linux; U; Android 2.2; en-gb; GT-P1000 Build/FROYO) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",

	//Firefox火狐
	"Mozilla/5.0 (Android; Mobile; rv:14.0) Gecko/14.0 Firefox/14.0",
	"Mozilla/5.0 (Android; Tablet; rv:14.0) Gecko/14.0 Firefox/14.0",
	"Mozilla/5.0 (Windows NT 6.2; WOW64; rv:21.0) Gecko/20100101 Firefox/21.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.8; rv:21.0) Gecko/20100101 Firefox/21.0",
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:21.0) Gecko/20130331 Firefox/21.0",

	//Google chrome
	"Mozilla/5.0 (Linux; Android 4.0.4; Galaxy Nexus Build/IMM76B) AppleWebKit/535.19 (KHTML, like Gecko) Chrome/18.0.1025.133 Mobile Safari/535.19",
	"Mozilla/5.0 (Linux; Android 4.1.2; Nexus 7 Build/JZ054K) AppleWebKit/535.19 (KHTML, like Gecko) Chrome/18.0.1025.166 Safari/535.19",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/27.0.1453.93 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/535.11 (KHTML, like Gecko) Ubuntu/11.10 Chromium/27.0.1453.93 Chrome/27.0.1453.93 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/27.0.1453.94 Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 6_1_4 like Mac OS X) AppleWebKit/536.26 (KHTML, like Gecko) CriOS/27.0.1453.10 Mobile/10B350 Safari/8536.25",

	//Internet Explore
	"Mozilla/5.0 (compatible; WOW64; MSIE 10.0; Windows NT 6.2)", //IE10
	"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0)", //IE9
	"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0)", //IE8
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.0)", //IE7
	"Mozilla/4.0 (Windows; MSIE 6.0; Windows NT 5.2)", //IE6

	//Opera
	"Opera/9.80 (Macintosh; Intel Mac OS X 10.6.8; U; en) Presto/2.9.168 Version/11.52", //Mac
	"Opera/9.80 (Windows NT 6.1; WOW64; U; en) Presto/2.10.229 Version/11.62", //Windows

	//Safari
	"Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_6; en-US) AppleWebKit/533.20.25 (KHTML, like Gecko) Version/5.0.4 Safari/533.20.27", //Mac
	"Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US) AppleWebKit/533.20.25 (KHTML, like Gecko) Version/5.0.4 Safari/533.20.27", //windows
	"Mozilla/5.0 (iPad; CPU OS 5_0 like Mac OS X) AppleWebKit/534.46 (KHTML, like Gecko) Version/5.1 Mobile/9A334 Safari/7534.48.3", //iPad
	"Mozilla/5.0 (iPhone; CPU iPhone OS 5_0 like Mac OS X) AppleWebKit/534.46 (KHTML, like Gecko) Version/5.1 Mobile/9A334 Safari/7534.48.3", //iPhone

	//iOS
	"Mozilla/5.0 (iPad; CPU OS 5_0 like Mac OS X) AppleWebKit/534.46 (KHTML, like Gecko) Version/5.1 Mobile/9A334 Safari/7534.48.3", //iPad
	"Mozilla/5.0 (iPhone; CPU iPhone OS 5_0 like Mac OS X) AppleWebKit/534.46 (KHTML, like Gecko) Version/5.1 Mobile/9A334 Safari/7534.48.", //iPhone
	"Mozilla/5.0 (iPod; U; CPU like Mac OS X; en) AppleWebKit/420.1 (KHTML, like Gecko) Version/3.0 Mobile/3A101a Safari/419.3", //iPod

	//Windows Phone
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows Phone OS 7.0; Trident/3.1; IEMobile/7.0; LG; GW910)", //windows Phone 7
	"Mozilla/5.0 (compatible; MSIE 9.0; Windows Phone OS 7.5; Trident/5.0; IEMobile/9.0; SAMSUNG; SGH-i917)",// Windows Phone7.5
	"	Mozilla/5.0 (compatible; MSIE 10.0; Windows Phone 8.0; Trident/6.0; IEMobile/10.0; ARM; Touch; NOKIA; Lumia 920)", //windows phone 8
}

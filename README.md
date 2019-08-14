## 以Golang实现的一种实时流量统计分析服务

#### 目前只实现了UV和PV的统计，其他业务统计分析需求再加入就行了 
 
#### 1) /src/run/run.go文件，用于模拟海量用户访问web网站，js打点nginx服务器响应体返回1px的像素图片，并且实时记录相应数据到log日志中的行为，--total参数，可以自由设置访问人数。

#### 2) /src/logs/dig.log是用户数据日志， runtime.log是运行时发生异常或其他错误信息的日志--日志等级在error或info层面。

#### 3) /src/analysis/analysis.go是多协程统计监控并分析日志信息的文件，可设置通过命令行 "--routineNum"参数并发度，默认并发度是5。

#### 4)整个项目的架构是使用CSP通信模型去做的并发任务：默认设置的是1个goroutine去做逐行读取dig.log日志的内容，5个goroutine通过channel并发地消费每行日志内容，1个做PV的goroutine和1个做UV的goroutine，分别通过2个channel与5个消费日志的goroutine进行通信，PV和UV的goroutine通过自己的channel与storage存储goroutine进行通信，storage存储goroutine把数据存储在redis的Hyperloglog的数据结构中，并做了相应的持久化。

#### 5)统计分析的是我当时自己为公司做的官网，地址是：https://github.com/GeekCoffee/GeekTech_Website

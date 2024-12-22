# golangBot
golang实现的微信机器人

## thanks
本项目基于[openwechat](https://openwechat.readthedocs.io/zh/latest/bot.html)实现
## 目前已经实现
1. 获取并发送摸鱼图片
2. 接入天气接口
3. 解析抖音视频链接并发送视频
## 计划实现
1. 定时任务自动推送消息
2. 支持指令开启或关闭机器人


```
cmd/
存放程序的入口点
通常只需要调用 internal/bot 包的功能
不包含具体业务逻辑
internal/bot/
机器人的核心配置
处理登录、初始化等基础功能
一般不需要修改，除非要改变机器人的基础行为
internal/handlers/
消息处理的路由层
决定不同消息应该调用哪个服务
添加新功能时需要在这里添加新的处理逻辑
internal/models/
定义数据结构
每个外部API的响应结构
添加新功能时，如果需要处理新的数据结构，就在这里添加
internal/services/
实现具体的业务逻辑
每个功能一个文件
处理API调用、数据处理等
添加新功能时，主要的代码都在这里
添加新功能的步骤：
先在 models/ 中定义需要的数据结构
在 services/ 中实现具体功能
在 handlers/message.go 中添加对应的处理逻辑
```
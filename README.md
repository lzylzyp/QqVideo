# QqVideo
腾讯视频VIP自动签到获取V力值

- 每日自动签到(默认执行时间:10:10)
- 每日观看60分钟视频获取V力值(默认执行时间:20:10)
![img_2.png](img_2.png)

### 使用教程
#### config.ini 配置说明
```ini
[cookie]
#腾讯视频登陆后的cookie(获取方法在下面)
QqVideoCookie = ``

[email]
Host = smtp.exmail.qq.com
Port = 465
Username = 通知邮箱账号
Pass = 密码
NotifyEmail = 接收通知的email
```
#### 腾讯视频cookie获取的一种方法
```js
//网页登陆腾讯视频后,控制台,输入
document.cookie
```
![img.png](img.png)

#### 启动程序
```shell
docker-compose up -d
```
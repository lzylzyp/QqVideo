package engine

import (
	"QqVideo/config"
	"QqVideo/email"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const (
	LoginUrl = "https://access.video.qq.com/user/auth_refresh?vappid=11059694&vsecret=fdf61a6be0aad57132bc5cdf78ac30145b6cd2c1470b0cfe&type" +
		"=qq&g_tk=&g_vstk=363833841&g_actk=382330955&callback=jQuery19108399751114128995_1660375409396&_=1660375409397"
	SignUrl      = "https://vip.video.qq.com/fcgi-bin/comm_cgi?name=hierarchical_task_system&cmd=2"   //sign url
	Minutes60Url = "https://vip.video.qq.com/fcgi-bin/comm_cgi?name=spp_MissionFaHuo&cmd=4&task_id=1" //60 minutes
)

var (
	JsonReg                = regexp.MustCompile(`QZOutputJson=\((.*)\);`)
	FindCookieVuSessionReg = regexp.MustCompile(`vqq_vusession=([^;]*){1}`)
)

type Engine struct {
	videoCookie     string //login cookie
	vuSessionCookie string //签到cookie
}

// Run 启动签到
func (e *Engine) Run(params *Params) {
	e.videoCookie = params.Cookie
	if err := e.getVuSessionCookie(); err != nil {
		email.SendEmail(config.NotifyEmail, "cookie解析出错", err.Error())
		return
	}
	bytes, err := e.httpRequest(params.ReqUrl, e.vuSessionCookie, "https://m.v.qq.com", false)
	if err != nil {
		log.Println(err)
		email.SendEmail(config.NotifyEmail, "获取V力值请求失败", err.Error())
		return
	}
	score, err := e.withRes(&bytes)
	if err != nil {
		log.Println(err)
		email.SendEmail(config.NotifyEmail, params.WithResErrMsg, err.Error())
		return
	}
	if params.ScoreDefine > 0 { //提前设置好积分
		score = params.ScoreDefine
	}
	email.SendEmail(config.NotifyEmail,
		params.EmailSubject,
		fmt.Sprintf(params.NotifyMsg,
			score))
	log.Println("获得积分:", score)
}

//getVuSessionCookie get VuSession
func (e *Engine) getVuSessionCookie() error {
	findVuSession := FindCookieVuSessionReg.FindStringSubmatch(e.videoCookie)
	if len(findVuSession) != 2 {
		return errors.New("cookie错误")
	}
	//替换成sprint string
	vuSessionCookieSprint := strings.Replace(e.videoCookie, findVuSession[1], "%s", 1)
	vuSession, err := e.httpRequest(LoginUrl, e.videoCookie, "https://film.qq.com/", true)
	if err != nil {
		return err
	}
	//获取v力值cookie获取成功
	e.vuSessionCookie = fmt.Sprintf(vuSessionCookieSprint, string(vuSession))
	return nil
}

//处理V力值签到结果
func (e *Engine) withRes(res *[]byte) (int, error) {
	RegSub := JsonReg.FindStringSubmatch(string(*res))
	if len(RegSub) != 2 {
		return 0, errors.New(fmt.Sprintf("返回数据解析失败,返回数据:%s", string(*res)))
	}

	resJson := e.formatJson([]byte(RegSub[1])) //转换数据格式
	if resJson.Ret != 0 {                      //error
		return 0, errors.New(fmt.Sprintf("V力值获取失败,errCode:%d,errMsg:%s", resJson.Ret, resJson.Msg))
	}
	score, _ := strconv.Atoi(resJson.CheckinScore)
	return score, nil
}

//httpRequest network request
func (e *Engine) httpRequest(url, cookieStr, referer string, isGetVuSession bool) ([]byte, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("cookie", cookieStr)
	req.Header.Set("referer", referer)
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.83 Safari/537.36")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if isGetVuSession { //get vqq_vusession
		for _, cookie := range res.Cookies() {
			if cookie.Name == "vqq_vusession" {
				return []byte(cookie.Value), nil
			}
		}
		return nil, errors.New("登陆cookie失效")
	}
	//V力值签到
	return ioutil.ReadAll(res.Body)
}

type ResJson struct {
	Ret          int    `json:"ret"`
	CheckinScore string `json:"checkin_score"`
	Msg          string `json:"msg"`
}

//转换返回json
func (e *Engine) formatJson(bytes []byte) *ResJson {
	var jsonStr ResJson
	json.Unmarshal(bytes, &jsonStr)
	return &jsonStr
}

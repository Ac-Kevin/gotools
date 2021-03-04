package gtimer

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

//时间参数
type timerArgs struct {
	hour int //小时
	min  int //分钟
	sec  int //秒
}

/*
TimedExec 定时执行某个函数
传入参数说明:
	tArgs 时间字符串 hour:min:sec 不能为空且不能包含字符 '.' 和 '-'
				取值范围 hour[0,24)，min[0:60)，sec[0,60)
				多个则用逗号隔开 例："1:1:1" 或 "01:01:01" 和 "00:00:01,15:04:05,23:59:59"
	first 让程序运行马上执行一次函数f
	f 要执行的函数（无参）
使用说明demo:
	TimedExec("10:48:00,10:48:30,10:48:15", false, youMethod)
*/
func TimedExec(tArgs string, first bool, f func()) error {

	//解析传入时间字符串 为 时间参数实例切片
	timeArr, err := decodeToTimerArgs(tArgs)
	if err != nil {
		return err
	}

	if first {
		go f()
	}

	//循环切片 开启定时器
	for _, t := range timeArr {
		go timerHandle(t, f)
	}
	return nil
}

/*TimedExecWithParams 定时执行某个带参数函数(如：每天凌晨3点执行某个带参(可以是多个参数)函数)定时执行某个函数
传入参数说明:
	tArgs 时间字符串 hour:min:sec 不能为空且不能包含字符 '.' 和 '-'
				取值范围 hour[0,24)，min[0:60)，sec[0,60)
				多个则用逗号隔开 例："1:1:1" 或 "01:01:01" 和 "00:00:01,15:04:05,23:59:59"
	first 让程序运行马上执行一次函数f
	f 传入的带参数函数(可接收多个参数)
	params f函数要接收的参数(可传入多个参数)
使用说明demo:
	TimedExec("10:48:00,10:48:30,10:48:15", false, youMethod, "hello world")
开发者：
  LC 2019.03.25
修改日志：
	LC 2019.04.12 修改传入的函数为可带多个参数(任意类型)，传入的参数也为多个参数(任意类型)
	mm 2019.04.16 修改为支持多时间配置
*/
func TimedExecWithParams(tArgs string, first bool, f func(...interface{}), params ...interface{}) error {

	//解析传入时间字符串 为 时间参数实例切片
	timeArr, err := decodeToTimerArgs(tArgs)
	if err != nil {
		return err
	}

	if first {
		go f(params...)
	}

	//循环切片 开启定时器
	for _, t := range timeArr {
		go timerHandleWithParams(t, f, params...)
	}
	return nil
}

/*timerHandle 定时器
传入参数：t 时间参数实例
				 f 要执行的函数（无参）
*/
func timerHandle(t timerArgs, f func()) {
	var (
		now    time.Time   // 当前时间
		offset time.Time   // 偏移时间
		next   time.Time   // 下一次执行时间
		timer  *time.Timer // 定时
	)

	for {
		now = time.Now()

		// 定时每天执行
		next = time.Date(now.Year(), now.Month(), now.Day(), t.hour, t.min, t.sec, 0, now.Location())

		if next.Sub(now) <= 0 {
			// 如果时间差小于等于0，说明当天执行时间已经过了，则时间偏移24hour
			offset = now.Add(24 * time.Hour)
			next = time.Date(offset.Year(), offset.Month(), offset.Day(), t.hour, t.min, t.sec, 0, offset.Location())
		}

		fmt.Println("TimedExec下一次执行间隔时间:", next.Sub(now))
		timer = time.NewTimer(next.Sub(now))
		<-timer.C

		f()
	}
}

/*timerHandleWithParams 定时器
传入参数：t 时间参数实例
				 f 要执行的函数（带参 不限个数和类型）
				 params 函数f的参数集（不限个数和类型）
*/
func timerHandleWithParams(t timerArgs, f func(...interface{}), params ...interface{}) {
	var (
		now    time.Time   // 当前时间
		offset time.Time   // 偏移时间
		next   time.Time   // 下一次执行时间
		timer  *time.Timer // 定时
	)

	for {
		now = time.Now()

		// 定时每天执行
		next = time.Date(now.Year(), now.Month(), now.Day(), t.hour, t.min, t.sec, 0, now.Location())

		if next.Sub(now) <= 0 {
			// 如果时间差小于等于0，说明当天执行时间已经过了，则时间偏移24hour
			offset = now.Add(24 * time.Hour)
			next = time.Date(offset.Year(), offset.Month(), offset.Day(), t.hour, t.min, t.sec, 0, offset.Location())
		}

		fmt.Println("TimedExec下一次执行间隔时间:", next.Sub(now))
		timer = time.NewTimer(next.Sub(now))
		<-timer.C

		f(params...)
	}
}

/*decodeToTimerArgs 解析时间参数
传入参数：tArgs 时间字符串 hour:min:sec 不能为空且不能包含字符 '.' 和 '-'
				 取值范围 hour[0,24)，min[0:60)，sec[0,60)
				 多个则用逗号隔开 例："1:1:1" 或 "01:01:01" 和 "00:00:01,15:04:05,23:59:59"
返回参数：result 解析后的时间实例切片
				 err 解析错误信息
*/
func decodeToTimerArgs(tArgs string) (result []timerArgs, err error) {
	defer func() {
		if tErr := recover(); tErr != nil {
			err = fmt.Errorf("时间字符串格式错误：%s", tErr)
		}
	}()
	if tArgs == "" || strings.Contains(tArgs, "-") || strings.Contains(tArgs, ".") {
		panic(fmt.Sprintf("时间字符串不能为空，不能包含字符 '.' 和 '-'，tArgs='%s'", tArgs))
	}
	var item timerArgs
	var tmp []string
	for _, t := range strings.Split(tArgs, ",") {
		if t != "" {
			tmp = strings.Split(t, ":")
			if item.hour, err = strconv.Atoi(tmp[0]); err == nil { //提取hour
				if item.min, err = strconv.Atoi(tmp[1]); err == nil { //提取min
					item.sec, err = strconv.Atoi(tmp[2]) //提取sec
				}
			}
			//err为空 时间取值范围正确
			if err == nil && item.hour < 24 && item.min < 60 && item.sec < 60 {
				result = append(result, item) //拼接item到结果切片中
			} else {
				panic(fmt.Sprintf("超出取值范围(正整数)：hour[0,24)，min[0:60)，sec[0,60)，t='%s'", t))
			}
		}
	}
	return
}

/*
DelayExec 隔多长时间执行某个无参函数(当前程序运行后，隔多长时间执行某个无参函数)  注：调用该方法必须使用 go 例: go sbjbase.DelayExec(day30ms, false, verifyGbsDevsnIntegrity)
传入参数说明:
	sec  秒
	f 传入的无参数函数
	first 让程序运行马上执行一次函数f
使用说明demo:
	DelayExec(44, false, f) -> 每隔44秒后执行f
开发者：
  LC 2019.03.25
  	 2019.08.13 修改为异步执行
*/
func DelayExec(sec int, first bool, f func()) error {
	var (
		now   time.Time   // 当前时间
		next  time.Time   // 下一次执行时间
		timer *time.Timer // 定时
	)

	if sec <= 0 {
		return fmt.Errorf("sbjbase.DelayExec error: 传进来的时间参数需要大于0")
	}

	go func(fn func()) {
		if first {
			fn()
		}

		for {
			now = time.Now()
			next = now.Add(time.Duration(sec) * time.Second)
			//	fmt.Println("DelayExec下一次执行间隔时间:", next.Sub(now))
			timer = time.NewTimer(next.Sub(now))
			<-timer.C
			fn()
		}
	}(f)
	return nil
}

//Convert2TimeString 时间格式转换
//功能: 时间格式转换  int转2006-01-02 15:04:05
//参数： 传入：unixTime int时间数据    返回： string "2006-01-02 15:04:05"格式
func Convert2TimeString(unixTime int) string {
	return time.Unix(int64(unixTime), 0).Format("2006-01-02 15:04:05")
}

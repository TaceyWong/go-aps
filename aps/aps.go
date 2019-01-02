// go-aps: A Golang Job Scheduling Package.
// 不是并发安全的~

//实现链式语法的核心在于每次返回的对象都是一样的

package aps

import (
	"errors"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// default location
var loc = time.Local

// 缺省提示语言
var LANG = "EN" 

//ChangeLoc 改变location
func ChangeLoc(newLocation *time.Location) {
	loc = newLocation
}

// MAXJOBNUM 最大Job数量
const MAXJOBNUM = 10000

type Job struct {

	//暂停间隔
	interval uint64

	// 要运行的JOb函数
	jobFunc string
	// 时间单位
	unit string
	// 可选
	atTime string

	// 最后运行时间
	lastRun time.Time
	// 下次运行时间
	nextRun time.Time
	// 缓存下次运行时间和最后一次运行时间之间的时间间隔
	period time.Duration

	// 指定周几开始运行
	startDay time.Weekday

	//函数任务映射存储
	funcs map[string]interface{}

	// 函数和参数之间的映射存储
	fparams map[string]([]interface{})
}

// 创建时间间隔型任务.
func NewJob(intervel uint64) *Job {
	return &Job{
		intervel,
		"", "", "",
		time.Unix(0, 0),
		time.Unix(0, 0), 0,
		time.Sunday,
		make(map[string]interface{}),
		make(map[string]([]interface{})),
	}
}

// 如果立即运行返回True
// 下次运行时间比当前晚即可
func (j *Job) shouldRun() bool {
	return time.Now().After(j.nextRun)
}

//立即运行JOb并重新调度它
func (j *Job) run() (result []reflect.Value, err error) {
	f := reflect.ValueOf(j.funcs[j.jobFunc])
	params := j.fparams[j.jobFunc]
	if len(params) != f.Type().NumIn() {
		err = errors.New("the number of param is not adapted")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	j.lastRun = time.Now()
	j.scheduleNextRun()
	return
}

// 利用反射获取函数名称
func getFunctionName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf((fn)).Pointer()).Name()
}

// 指定要执行的目标Job函数和参数
func (j *Job) Do(jobFun interface{}, params ...interface{}) {
	typ := reflect.TypeOf(jobFun)
	if typ.Kind() != reflect.Func {
		panic("only function can be schedule into the job queue.")
	}

	fname := getFunctionName(jobFun)
	j.funcs[fname] = jobFun
	j.fparams[fname] = params
	j.jobFunc = fname
	//调度下一次运行
	j.scheduleNextRun()
}


// 格式化时间字符串（小时、分钟）
func formatTime(t string) (hour, min int, err error) {
	var er = errors.New("time format error")
	ts := strings.Split(t, ":")
	if len(ts) != 2 {
		err = er
		return
	}

	hour, err = strconv.Atoi(ts[0])
	if err != nil {
		return
	}
	min, err = strconv.Atoi(ts[1])
	if err != nil {
		return
	}

	if hour < 0 || hour > 23 || min < 0 || min > 59 {
		err = er
		return
	}
	return hour, min, nil
}

//	s.Every(1).Day().At("10:30").Do(task)
//	s.Every(1).Monday().At("10:30").Do(task)
func (j *Job) At(t string) *Job {
	hour, min, err := formatTime(t)
	if err != nil {
		panic(err)
	}

	// time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	mock := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), int(hour), int(min), 0, 0, loc)

	if j.unit == "days" {
		if time.Now().After(mock) {
			j.lastRun = mock
		} else {
			j.lastRun = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-1, hour, min, 0, 0, loc)
		}
	} else if j.unit == "weeks" {
		if j.startDay != time.Now().Weekday() || (time.Now().After(mock) && j.startDay == time.Now().Weekday()) {
			i := mock.Weekday() - j.startDay
			if i < 0 {
				i = 7 + i
			}
			j.lastRun = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-int(i), hour, min, 0, 0, loc)
		} else {
			j.lastRun = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-7, hour, min, 0, 0, loc)
		}
	}
	return j
}

//Compute the instant when this job should run next
func (j *Job) scheduleNextRun() {
	if j.lastRun == time.Unix(0, 0) {
		if j.unit == "weeks" {
			i := time.Now().Weekday() - j.startDay
			if i < 0 {
				i = 7 + i
			}
			j.lastRun = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-int(i), 0, 0, 0, 0, loc)

		} else {
			j.lastRun = time.Now()
		}
	}

	if j.period != 0 {
		// translate all the units to the Seconds
		j.nextRun = j.lastRun.Add(j.period * time.Second)
	} else {
		switch j.unit {
		case "minutes":
			j.period = time.Duration(j.interval * 60)
			break
		case "hours":
			j.period = time.Duration(j.interval * 60 * 60)
			break
		case "days":
			j.period = time.Duration(j.interval * 60 * 60 * 24)
			break
		case "weeks":
			j.period = time.Duration(j.interval * 60 * 60 * 24 * 7)
			break
		case "seconds":
			j.period = time.Duration(j.interval)
		}
		j.nextRun = j.lastRun.Add(j.period * time.Second)
	}
}

// NextScheduledTime returns the time of when this job is to run next
func (j *Job) NextScheduledTime() time.Time {
	return j.nextRun
}

// 设置时间单位
func (j *Job) Second() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	job = j.Seconds()
	return
}

// 设置时间单位
func (j *Job) Seconds() (job *Job) {
	j.unit = "seconds"
	return j
}

// 设置时间单位
func (j *Job) Minute() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	job = j.Minutes()
	return
}

//设置时间单位
func (j *Job) Minutes() (job *Job) {
	j.unit = "minutes"
	return j
}

//设置时间单位
func (j *Job) Hour() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	job = j.Hours()
	return
}

// 设置时间单位
func (j *Job) Hours() (job *Job) {
	j.unit = "hours"
	return j
}

// 设置时间单位
func (j *Job) Day() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	job = j.Days()
	return
}

// 设置时间单位
func (j *Job) Days() *Job {
	j.unit = "days"
	return j
}

// s.Every(1).Monday().Do(task)
// 设置开始时间星期一
func (j *Job) Monday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 1
	job = j.Weeks()
	return
}

// 设置开始时间星期二
func (j *Job) Tuesday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 2
	job = j.Weeks()
	return
}

// 设置开始时间星期三
func (j *Job) Wednesday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 3
	job = j.Weeks()
	return
}

// 设置开始时间星期四
func (j *Job) Thursday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 4
	job = j.Weeks()
	return
}

// 设置开始时间星期五
func (j *Job) Friday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 5
	job = j.Weeks()
	return
}

// 设置开始时间星期六
func (j *Job) Saturday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 6
	job = j.Weeks()
	return
}

// 设置开始时间星期日
func (j *Job) Sunday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 0
	job = j.Weeks()
	return
}

//设置时间单位为周
func (j *Job) Weeks() *Job {
	j.unit = "weeks"
	return j
}

// Class Scheduler, the only data member is the list of jobs.
// 调度器，唯一的数据成员为job列表
type Scheduler struct {
	// 存储Job的数组，预先定义好了可执行多少任务
	jobs [MAXJOBNUM]*Job

	// 存有的Job数据
	size int
}

// 实现排序接口，根据下次运行时间
func (s *Scheduler) Len() int {
	return s.size
}

func (s *Scheduler) Swap(i, j int) {
	s.jobs[i], s.jobs[j] = s.jobs[j], s.jobs[i]
}

func (s *Scheduler) Less(i, j int) bool {
	//排序判断条件：nextRun
	return s.jobs[j].nextRun.After(s.jobs[i].nextRun)
}

// 创建一个新的调度器
func NewScheduler() *Scheduler {
	return &Scheduler{[MAXJOBNUM]*Job{}, 0}
}

// 获取当前可运行的Job。判断条件是shouldRun返回True

func (s *Scheduler) getRunnableJobs() (running_jobs [MAXJOBNUM]*Job, n int) {
	runnableJobs := [MAXJOBNUM]*Job{}
	n = 0
	sort.Sort(s)
	for i := 0; i < s.size; i++ {
		if s.jobs[i].shouldRun() {

			runnableJobs[n] = s.jobs[i]
			//fmt.Println(runnableJobs)
			n++
		} else {
			break
		}
	}
	return runnableJobs, n
}

// 返回下个要运行Job的时间
func (s *Scheduler) NextRun() (*Job, time.Time) {
	if s.size <= 0 {
		return nil, time.Now()
	}
	sort.Sort(s)
	return s.jobs[0], s.jobs[0].nextRun
}

// Schedule a new periodic job
func (s *Scheduler) Every(interval uint64) *Job {
	job := NewJob(interval)
	s.jobs[s.size] = job
	s.size++
	return job
}

// Run all the jobs that are scheduled to run.
func (s *Scheduler) RunPending() {
	runnableJobs, n := s.getRunnableJobs()

	if n != 0 {
		for i := 0; i < n; i++ {
			runnableJobs[i].run()
		}
	}
}

// Run all jobs regardless if they are scheduled to run or not
func (s *Scheduler) RunAll() {
	for i := 0; i < s.size; i++ {
		s.jobs[i].run()
	}
}

// Run all jobs with delay seconds
func (s *Scheduler) RunAllwithDelay(d int) {
	for i := 0; i < s.size; i++ {
		s.jobs[i].run()
		time.Sleep(time.Duration(d))
	}
}

// Remove specific job j
func (s *Scheduler) Remove(j interface{}) {
	i := 0
	//左index
	for ; i < s.size; i++ {
		if s.jobs[i].jobFunc == getFunctionName(j) {
			break
		}
	}

	// 指定目标job后面的job全部往前移
	for j := (i + 1); j < s.size; j++ {
		s.jobs[i] = s.jobs[j]
		i++
	}
	s.size = s.size - 1
}

// 清除所有调度的任务
func (s *Scheduler) Clear() {
	//数组的每一个元素置空
	for i := 0; i < s.size; i++ {
		s.jobs[i] = nil
	}
	// size置零
	s.size = 0
}

// 开始调度
func (s *Scheduler) Start() chan bool {
	stopped := make(chan bool, 1)
	ticker := time.NewTicker(1 * time.Second)
	//开启调度，每一秒检查一次
	go func() {
		for {
			select {
			case <-ticker.C:
				s.RunPending()
			case <-stopped:
				return
			}
		}
	}()

	return stopped
}

// The following methods are shortcuts for not having to
// create a Schduler instance

var defaultScheduler = NewScheduler()
var jobs = defaultScheduler.jobs

// Schedule a new periodic job
func Every(interval uint64) *Job {
	return defaultScheduler.Every(interval)
}

// Run all jobs that are scheduled to run
//
// Please note that it is *intended behavior that run_pending()
// does not run missed jobs*. For example, if you've registered a job
// that should run every minute and you only call run_pending()
// in one hour increments then your job won't be run 60 times in
// between but only once.
func RunPending() {
	defaultScheduler.RunPending()
}

// Run all jobs regardless if they are scheduled to run or not.
func RunAll() {
	defaultScheduler.RunAll()
}

// Run all the jobs with a delay in seconds
//
// A delay of `delay` seconds is added between each job. This can help
// to distribute the system load generated by the jobs more evenly over
// time.
func RunAllwithDelay(d int) {
	defaultScheduler.RunAllwithDelay(d)
}

// 运行所有调度器中的Job
func Start() chan bool {
	return defaultScheduler.Start()
}

// Clear 清空调度器中的所有Job
func Clear() {
	defaultScheduler.Clear()
}

// Remove 移除某个任务
func Remove(j interface{}) {
	defaultScheduler.Remove(j)
}

// NextRun 获取下次要运行的Job及时间
func NextRun() (job *Job, time time.Time) {
	return defaultScheduler.NextRun()
}
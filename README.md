# GO-APS

主要组件：

+ triggers:触发器
+ job stores：任务存储
+ executors：执行器
+ schedulers：调度器

其中：

调度器支持：

+ BlockingScheduler:当调度器是你的唯一进程
+ BackgroundScheduler: 在你的程序的内部后台运行

Job存储：

+ 内存：不持久化，重启应用失效
+ 磁盘文件
+ Sqlite
+ Redis
+ Mongo
+ MySQL
+ Postgresql

执行器：

+ 设置Go程最大数量

触发器：

+ date：在一个特定的时间点执行一次
+ interval：间隔时间执行
+ cron：在特定的时间执行

也可以把多种触发器结合使用




##### linux core tcp connection optimization

###### linux tcp ipv4 settings
 - `fs.file-max`
    max open file discriptor.
 - `net.ipv4.tcp_max_syn_backlog`
    max tcp syn queue length. tcp半连接队列最大长度。Tcp syn队列的最大长度，在进行系统调用connect时会发生Tcp的三次握手，server内核会为Tcp维护两个队列，Syn队列和Accept队列，Syn队列是指存放完成第一次握手的连接，Accept队列是存放完成整个Tcp三次握手的连接，修改net.ipv4.tcp_max_syn_backlog使之增大可以接受更多的网络连接。
    注意此参数过大可能遭遇到Syn flood攻击，即对方发送多个Syn报文端填充满Syn队列，使server无法继续接受其他连接。
    自 Linux 内核 2.2 版本以后，backlog 为已完成连接队列的最大值，未完成连接队列大小以 /proc/sys/net/ipv4/tcp_max_syn_backlog 确定，但是已连接队列大小受 SOMAXCONN 限制，为 min(backlog, SOMAXCONN)。
 - `net.ipv4.tcp_syncookies`
    修改此参数可以有效的防范上面所说的syn flood攻击

    原理：在Tcp服务器收到Tcp Syn包并返回Tcp Syn+ack包时，不专门分配一个数据区（不加入半连接队列），而是根据这个Syn包计算出一个cookie值。在收到Tcp ack包时，Tcp服务器在根据那个cookie值检查这个Tcp ack包的合法性。如果合法，再分配专门的数据区进行处理未来的TCP连接（加入全连接队列）。

    默认为0，1表示开启
    https://bbs.huaweicloud.com/blogs/195697
  - `net.ipv4.tcp_keepalive_time`
    Tcp keepalive心跳包机制，用于检测连接是否已断开，我们可以修改默认时间来间断心跳包发送的频率。

    keepalive一般是服务器对客户端进行发送查看客户端是否在线，因为服务器为客户端分配一定的资源，但是Tcp 的keepalive机制很有争议，因为它们可耗费一定的带宽。

    Tcp keepalive详情见Tcp/ip详解卷1 第23章

  - `net.ipv4.tcp_tw_reuse`
    我的上一篇文章中写到了time_wait状态，大量处于time_wait状态是很浪费资源的，它们占用server的描述符等。

    修改此参数，允许重用处于time_wait的socket。

    默认为0，1表示开启
  - `net.ipv4.tcp_tw_recycle`
    也是针对time_wait状态的，该参数表示快速回收处于time_wait的socket。

    默认为0，1表示开启
  - `net.ipv4.tcp_fin_timeout`
    修改time_wait状的存在时间，默认的2MSL（报文最大生存）时间
  - `net.ipv4.tcp_max_tw_buckets`
    所允许存在time_wait状态的最大数值，超过则立刻被清楚并且警告。

  - `net.ipv4.tcp_synack_retries`
    三次握手中，在第一步server收到client的syn后，把这个连接信息放到半连接队列中，同时回复syn+ack给client（第二步）；

    第三步的时候server收到client的ack，如果这时全连接队列没满，那么从半连接队列拿出这个连接的信息放入到全连接队列中，否则按tcp_abort_on_overflow指示的执行。

    这时如果全连接队列满了并且tcp_abort_on_overflow是0的话，server过一段时间再次发送syn+ack给client（也就是重新走握手的第二步），如果client超时等待比较短，client就很容易异常了。

    在我们的os中retry 第二步的默认次数是2（centos默认是5次）：
  - `net.ipv4.ip_local_port_range`
    表示对外连接的端口范围。
  - `somaxconn`
    前面说了Syn队列的最大长度限制，somaxconn参数决定Accept队列长度，在listen函数调用时backlog参数即决定Accept队列的长度，该参数太小也会限制最大并发连接数，因为同一时间完成3次握手的连接数量太小，server处理连接速度也就越慢。服务器端调用accept函数实际上就是从已连接Accept队列中取走完成三次握手的连接。

    Accept队列和Syn队列是listen函数完成创建维护的。

    /proc/sys/net/core/somaxconn修改

##### linux tcp debug command

- `netstat -s | grep "SYNstoLISTEN"`
    查看半连接队列溢出次数

- `netstat -s | grep "listen|LISTEN"`
    查看全连接队列溢出次数
- `ss -lnt`
    查看listen端口上的全连接队列最大长度，和当前长度
    全连接队列长度取决于 min(backlog, somaxconn)
    半连接最大长度取决于 max(64,  /proc/sys/net/ipv4/tcp_max_syn_backlog)
- `netstat -tn`
    netstat跟ss命令一样也能看到Send-Q、Recv-Q这些状态信息，不过如果这个连接不是Listen状态的话，Recv-Q就是指收到的数据还在缓存中，还没被进程读取，这个值就是还没被进程读取的 bytes；而 Send 则是发送队列中没有被远程主机确认的 bytes 数。
- `netstat -n | awk '/^tcp/ {++S[$NF]} END{for(a in S) print a, S[a]}'`
    结果如下：
    ```shell
    LAST_ACK 14
    SYN_RECV 348
    ESTABLISHED 70
    FIN_WAIT1 229
    FIN_WAIT2 30
    CLOSING 33
    TIME_WAIT 18122
    ```
    命令中的含义分别如下。

    CLOSED：无活动的或正在进行的连接。
    LISTEN：服务器正在等待进入呼叫。
    SYN_RECV：一个连接请求已经到达，等待确认。
    SYN_SENT：应用已经开始，打开一个连接。
    ESTABLISHED：正常数据传输状态。
    FIN_WAIT1：应用说它已经完成。
    FIN_WAIT2：另一边已同意释放。
    ITMED_WAIT：等待所有分组死掉。
    CLOSING：两边尝试同时关闭。
    TIME_WAIT：另一边已初始化一个释放。
    LAST_ACK：等待所有分组死掉。


##### linux tcp connection optimization

- 一般配置
    `vim /etc/sysctl.conf`
    
    ```shell
    net.ipv4.tcp_fin_timeout = 30
    net.ipv4.tcp_keepalive_time = 1200
    net.ipv4.tcp_syncookies = 1
    net.ipv4.tcp_tw_reuse = 1
    net.ipv4.tcp_tw_recycle = 1
    net.ipv4.ip_local_port_range = 10000 65000
    net.ipv4.tcp_max_syn_backlog = 8192
    net.ipv4.tcp_max_tw_buckets = 5000
    ```

    net.ipv4.tcp_syncookies＝1表示开启SYN Cookies。当出现SYN等待队列溢出时，启用Cookie来处理，可防范少量的SYN攻击。该参数默认为0，表示关闭。
    net.ipv4.tcp_tw_reuse＝1表示开启重用，即允许将TIME-WAIT套接字重新用于新的TCP连接。该参数默认为0，表示关闭。
    net.ipv4.tcp_tw_recycle＝1表示开启TCP连接中TIME-WAIT套接字的快速回收，该参数默认为0，表示关闭。
    net.ipv4.tcp_fin_timeout＝30表示如果套接字由本端要求关闭，那么这个参数将决定它保持在FIN-WAIT-2状态的时间。
    net.ipv4.tcp_keepalive_time＝1200表示当Keepalived启用时，TCP发送Keepalived消息的频度改为20分钟，默认值是2小时。
    net.ipv4.ip_local_port_range＝1000065000表示CentOS系统向外连接的端口范围。其默认值很小，这里改为10000到65000。建议不要将这里的最低值设得太低，否则可能会占用正常的端口。
    net.ipv4.tcp_max_syn_backlog＝8192表示SYN队列的长度，默认值为1024，此处加大队列长度为8192，可以容纳更多等待连接的网络连接数。
    net.ipv4.tcp_max_tw_buckets＝5000表示系统同时保持TIME_WAIT套接字的最大数量，如果超过这个数字，TIME_WAIT套接字将立刻被清除并打印警告信息，默认值为180000，此处改为5000。对于Apache、Nginx等服务器，前面介绍的几个参数已经可以很好地减少TIME_WAIT套接字的数量，但是对于Squid来说，效果却不大，有了此参数就可以控制TIME_WAIT套接字的最大数量，避免Squid服务器被大量的TIME_WAIT套接字拖死。
    执行以下命令使内核配置立马生效：
    `/sbin/sysctl –p`

- nginx服务器配置
    ```shell
    net.ipv4.tcp_syncookies=1
    net.ipv4.tcp_tw_reuse=1
    net.ipv4.tcp_tw_recycle = 1
    net.ipv4.ip_local_port_range = 10000 65000
    ```

- 邮件服务器

    ```shell
    net.ipv4.tcp_fin_timeout = 30
    net.ipv4.tcp_keepalive_time = 300
    net.ipv4.tcp_tw_reuse = 1
    net.ipv4.tcp_tw_recycle = 1
    net.ipv4.ip_local_port_range = 10000 65000
    kernel.shmmax = 134217728
    ```

- 其他配置

    ```shell
    #表示开启重用。允许将TIME-WAIT sockets重新用于新的TCP连接，默认为0，表示关闭；
    net.ipv4.tcp_syncookies = 1
    
    #一个布尔类型的标志，控制着当有很多的连接请求时内核的行为。启用的话，如果服务超载，内核将主动地发送RST包。
    net.ipv4.tcp_abort_on_overflow = 1
    
    #表示系统同时保持TIME_WAIT的最大数量，如果超过这个数字，TIME_WAIT将立刻被清除并打印警告信息。
    #默认为180000，改为6000。对于Apache、Nginx等服务器，此项参数可以控制TIME_WAIT的最大数量,服务器被大量的TIME_WAIT拖死
    net.ipv4.tcp_max_tw_buckets = 6000
    
    #有选择的应答
    net.ipv4.tcp_sack = 1
    
    #该文件表示设置tcp/ip会话的滑动窗口大小是否可变。参数值为布尔值，为1时表示可变，为0时表示不可变。tcp/ip通常使用的窗口最大可达到65535 字节，对于高速网络.
    #该值可能太小，这时候如果启用了该功能，可以使tcp/ip滑动窗口大小增大数个数量级，从而提高数据传输的能力。
    net.ipv4.tcp_window_scaling = 1
    
    #TCP接收缓冲区
    net.ipv4.tcp_rmem = 4096        87380  4194304
    
    #TCP发送缓冲区
    net.ipv4.tcp_wmem = 4096        66384  4194304
    
    #Out of socket memory
    net.ipv4.tcp_mem = 94500000 915000000 927000000
    
    #该文件表示每个套接字所允许的最大缓冲区的大小。
    net.core.optmem_max = 81920
    
    #该文件指定了发送套接字缓冲区大小的缺省值（以字节为单位）。
    net.core.wmem_default = 8388608
    
    #指定了发送套接字缓冲区大小的最大值（以字节为单位）。
    net.core.wmem_max = 16777216
    
    #指定了接收套接字缓冲区大小的缺省值（以字节为单位）。
    net.core.rmem_default = 8388608
    
    
    #指定了接收套接字缓冲区大小的最大值（以字节为单位）。
    net.core.rmem_max = 16777216
    
    
    #表示SYN队列的长度,默认为1024,加大队列长度为10200000,可以容纳更多等待连接的网络连接数。
    net.ipv4.tcp_max_syn_backlog = 1020000
    
    
    #每个网络接口接收数据包的速率比内核处理这些包的速率快时，允许送到队列的数据包的最大数目。
    net.core.netdev_max_backlog = 862144
    
    
    #web 应用中listen 函数的backlog 默认会给我们内核参数的net.core.somaxconn 限制到128，而nginx 定义的NGX_LISTEN_BACKLOG 默认为511，所以有必要调整这个值。
    net.core.somaxconn = 262144
    
    
    #系统中最多有多少个TCP 套接字不被关联到任何一个用户文件句柄上。如果超过这个数字，孤儿连接将即刻被复位并打印出警告信息。
    #这个限制仅仅是为了防止简单的DoS 攻击，不能过分依靠它或者人为地减小这个值，更应该增加这个
    net.ipv4.tcp_max_orphans = 327680
    
    #时间戳可以避免序列号的卷绕。一个1Gbps 的链路肯定会遇到以前用过的序列号。时间戳能够让内核接受这种“异常”的数据包。这里需要将其关掉。
    net.ipv4.tcp_timestamps = 0
    
    
    #为了打开对端的连接，内核需要发送一个SYN 并附带一个回应前面一个SYN 的ACK。也就是所谓三次握手中的第二次握手。这个设置决定了内核放弃连接之前发送SYN+ACK 包的数量。
    net.ipv4.tcp_synack_retries = 1
    
    
    #在内核放弃建立连接之前发送SYN 包的数量。
    net.ipv4.tcp_syn_retries = 1
    
    #表示开启重用。允许将TIME-WAIT sockets重新用于新的TCP连接，默认为0，表示关闭；https://zhuanlan.zhihu.com/p/40013724
    net.ipv4.tcp_tw_reuse = 1
    
    #修改系統默认的 TIMEOUT 时间。
    net.ipv4.tcp_fin_timeout = 15
    
    #表示当keepalive起用的时候，TCP发送keepalive消息的频度。缺省是2小时，建议改为20分钟。
    net.ipv4.tcp_keepalive_time = 30
    
    #表示用于向外连接的端口范围。缺省情况下很小：32768到61000，改为1024到65535。（注意：这里不要将最低值设的太低，否则可能会占用掉正常的端口！）
    net.ipv4.ip_local_port_range = 1024    65535
    
    #以下可能需要加载ip_conntrack模块 modprobe ip_conntrack ,有文档说防火墙开启情况下此模块失效
    #縮短established的超時時間
    net.netfilter.nf_conntrack_tcp_timeout_established = 180
    
    
    #CONNTRACK_MAX 允许的最大跟踪连接条目，是在内核内存中netfilter可以同时处理的“任务”（连接跟踪条目）
    net.netfilter.nf_conntrack_max = 1048576
    net.nf_conntrack_max = 1048576
    ```

    添加完成后 执行如下命令
    ```shell
    /sbin/sysctl -p /etc/sysctl.conf
    /sbin/sysctl -w net.ipv4.route.flush=1
    ```

##### TFO (TCP fast open)

  - `net.ipv4.tcp_fastopen = 3`
  由于只有客户端和服务器同时支持时，TFO 功能才能使用，所以 tcp_fastopen 参数是按比特位控制的。其中，第1 个比特位为 1 时，表示作为客户端时支持 TFO；第 2 个比特位为 1 时，表示作为服务器时支持 TFO，所以当 tcp_fastopen 的值为 3 时（比特为 0x11）就表示完全支持 TFO 功能。



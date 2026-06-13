# 文档说明

# 目录
- [前言](#前言)
- [鉴权系统](#鉴权系统)
    - [AdminAuth](#AdminAuth)
    - [TokenAuth](#TokenAuth)
    - [PosixAuth](#PosixAuth)
- [协议设计](#协议设计)
    - [AdminAuth协议接口](#AdminAuth协议接口)
    - [TokenAuth协议接口](#TokenAuth协议接口)
    - [PosixAuth协议接口](#PosixAuth协议接口)

## 前言
因使用nfs访问方式存在局域网存储访问隐患，故重写了存储鉴权结构以增强存储访问的安全性。

存储安全分物理存储安全及软件存储安全两大部份。

物理存储安全需要依靠上层应用做多副本存储，
本节点中只提供单机型的raid6级的raid冗余，最大可损坏两块硬盘。

软件存储安全主要为访问安全，分本地访问与网络访问两部分。
本地访问依赖于linux本身的安全；本设计主要针对网络访问安全而设计。

网络访问主要参考了oauth2的鉴权流程对外提供文件增、删、查、改功能，
并根据filecoin的系统文件要求提供对应的posix规范综合进行了以下改造:

## 鉴权系统
```
鉴权系统主要提供了基本安全密钥生成，包括两部分：
AdminAuth 基本鉴权，需要通过https接入, 在miner机上进行管理。
TokenAuth 会话授权，通过AdminAuth换取每一次会话授权所需要的token，
          此权限可以通过http进行文件的增、删、查、改功能。 
PosixAuth(TODO) Posix接口，因Filecoin需要Posix文件访问，
          此授权通过AdminAuth生成只读的token仅授权给miner、wdpost访问，此访问应受AdminAuth管理。
```

### AdminAuth
```
存储节点初始化时，会有一个初始的密钥。

当存储节点被添加到miner节点时，miner通过初始密钥对存储节点的AdminAuth进行接管；

miner接管理AdminAuth后，重新重启miner会进行一次密钥变更，
同时提供了手工变更AdminAuth密钥的功能，防止AdminAuth被强破的可能性。

miner管理存储节点的的AdminAuth时，每一台存储节点会一个唯一不同的密钥，
不同的存储节点需要不同的AdminAuth访问。

miner需要被视为可信的物理节点，当miner节点不可信时，应及时变更存储节点密码，
默认情况下重启miner即可变，不需人工介入。
```

### TokenAuth
```
此为Filecoin的每一个扇区访问安全而设计，即不同扇区每一次会话会有不同的TokenAuth，
以确保每一个扇区的操作使用的是唯一密钥。

当需要操作存储节点上的某个扇区(上传、下载、读取、删除)时，需通过miner获取该扇区的会话授权token，
使用完后自行释放，不释放时默认一定时后会自行释放。

此设计专为大文件传输而设计，小文件请使用PoxixAuth接口。
```

### PosixAuth
```
go-nfs, go-fuse在golang的文件大规模读挂载中测试未通过，因此仍采用原生nfs做为挂载接口，但为只读。
```

## 协议设计

### AdminAuth协议接口
```
指令接口，走https协议
/check -- 监控zfs磁盘状态，等价于zfs status -x, 只读
/sys/auth/change?token=d41d8cd98f00b204e9800998ecf8427e 首次使用时须调此接口重置密钥，成功时返回200及新密钥
/sys/file/token?file=filepath -- 获取临时的token锁
```


### TokenAuth协议接口
```
文件传输专用，因大文件而走http协议
/file/capacity -- 获取用户空间容量
/file/move?file=filepath&new=filepath -- 重命名
/file/delete?filepath=filepath -- 删除路径，TODO:该操作将只是临时删除，7天后永久性删除。
/file/upload?file=xxx&pos=0&checksum=sha1 -- POST, 文件上传，因大文件而走http，但需要填写BaseAuth的username为s-f0xxx-xx,password为临时的token;checksum当前固定为sha1,其他值不返回hash值, 内置支持文件夹上传与自动断点续传功作
/file/list?file=xxx -- 列出文件的信息
/file/download?file=xxx -- 文件读取, 同文件服务器，pos与checksum选填
```

### PosixAuth协议接口

go-nfs, go-fuse在golang的文件大规模读挂载中测试未通过，因此仍采用原生nfs做为挂载接口，但只读。


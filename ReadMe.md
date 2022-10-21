### ghosts
----------------
获取dns并自动修改和备份hosts文件，加速github访问。

支持Linux、MacOS、Windows、FreeBSD。

### 背景
目前github加速访问，主要有以下几种方式：

1、修改hosts；
2、加速工具，如dev-sidecar, fastgithub，watt toolkit等；
3、github镜像；
4、梯子；
其中，现有的hosts修改方案总是建议用额外的hosts修改软件，加速工具多不支持ssh，镜像不稳定，梯子一般需要付费。

ghosts可作为加速工具的补充，只在需要的时候手动运行一次，就可以备份和修改hosts文件，从而加速github ssh模式下推拉代码。

### 使用方法
1、如果已经配置好go开发环境，则直接使用go install安装：

```shell
go install github.com/moqsien/ghosts

ghosts // ghost.exe on windows
```

2、二进制安装文件：

[下载](https://github.com/moqsien/ghosts/releases/tag/v0.0.2)

下载完成后，可以自行设置环境变量，或者将二进制文件移动到已在环境变量中的目录下。

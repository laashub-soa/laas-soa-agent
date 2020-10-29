自己写一个jenkins-构建器/gitlab-runner?



builder(构建器)

基于docker进行资源隔离和构建器管理

运行依赖于服务器

启动时指定服务端的连接信息(基于websocket?), 连接到服务器

服务端通过ssh连接到构建机器进行初始化构建器





构建过程为:

根据项目目录创建并切换到运行目录, 目录结构如下:

```
配置根目录
	dependency_lib
	git配置名称
		仓库目录
			branch
				master
					source
                    build
                    target
				xxx
			tag
				xxx
					source
                    build
                    target

```

source中存源码, build中存构建脚本文件, 使用version文件记录构建脚本文件的版本, 当版本一致时则不拉取最新构建脚本, 反之则拉取最新构建脚本

拉取源码

​	tag中的代码不会缓存, branch中会缓存最近一次代码, 首次使用完整clone, 第二次使用fetch

创建容器进行构建

​	找寻`build_依赖镜像`镜像是否存在, 如果不存在则根据依赖镜镜像构建`build_依赖镜像`镜像, 如果存在则启动镜像、传递参数、挂载target目录

​		拷贝编译脚本进去, 执行编译脚本

​	

服务端需要配置:

​	构建机器信息:

​		服务器连接信息: ip、port、username、password

​	镜像仓库信息:

​		描述信息: 名称、描述

​		连接信息: ip、port、username、password

​	父镜像

​			编译环境

​			编译脚本

​			打包镜像脚本

​	构建信息:

​			编译环境

​			编译脚本

​			打包镜像脚本

​	git服务器信息:

​			连接信息: ip、port、username、password, 该连接账号的权限只需要只读即可

​			是否默认

​	配置项目信息:

​		git服务器: 默认使用默认的git服务器

​		地址: 可从git服务器选择, 也可以直接输入

​		项目描述: 如果是gitlab服务器则自动带出项目描述, 也可以直接输入

​		项目类型: java/php/nodejs/python

​		构建信息: 选择一个构建信息, 可进行预览





和服务端的维持关系:

使用websocket方式

# 案例

之前基于gitlab-runner做ci的案例:

父镜像Dockerfile:

```
FROM apache/skywalking-base:6.5.0 as skywalking
FROM anapsix/alpine-java
COPY --from=skywalking /skywalking/agent /agent
ENV SW_AGENT_NAMESPACE='tristan' SW_AGENT_NAME=${IMAGE_PROJECT_TAG} SW_AGENT_COLLECTOR_BACKEND_SERVICES='skywalking-skywalking-oap.skywalking:11800'
RUN echo 'Asia/Shanghai' > /etc/timezone
```

被依赖项目的.gitlab-ci.yml:

```
build:code:
  image: maven:3-alpine
  variables:
    MAVEN_CLI_OPTS: "-s .m2/settings.xml --batch-mode"
    GIT_STRATEGY: clone
  script:
    - echo "Asia/Shanghai" > /etc/timezone
    - mvn deploy
```

部署型项目的.gitlab-ci.yml:

```
build:code:
  stage: build
  image: maven:3-alpine
  variables:
    MAVEN_CLI_OPTS: "-s .m2/settings.xml --batch-mode"
    GIT_STRATEGY: clone
  cache:
    paths:
      - target/
  script:
    - mvn clean package -DskipTests
test:image:
  stage: test
  image: docker
  cache:
    paths:
      - target/
  dependencies:
    - :build:code
  script:
    - chmod 777 build-docker.sh && dos2unix build-docker.sh && source build-docker.sh
```

build-docker.sh:

```
#!/bin/bash
echo '>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>开始构建服务:'${JOB_NAME}
dos2unix /root/.m2/dockerregistry && source /root/.m2/dockerregistry
CUR_DATETIME_STR=$(date "+%Y%m%d%H%M")
IMAGE_PROJECT_TAG=${CI_PROJECT_NAME}"-"${CI_COMMIT_REF_NAME}
IMAGE_ID=${DOCKER_REGISTRY_URL}"/"${IMAGE_PROJECT_TAG}":"${CI_BUILD_ID}"_"${CUR_DATETIME_STR}
echo '>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>正在登录镜像仓库'
cat /root/.m2/dockerregistry-auth |  docker login ${DOCKER_REGISTRY_URL} --username ${DOCKER_REGISTRY_USERNAME} --password-stdin
echo '<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<登录镜像仓库成功'
echo '<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<正在构建镜像'
docker build --build-arg IMAGE_PROJECT_TAG=${IMAGE_PROJECT_TAG} -t ${IMAGE_ID}  .
echo '<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<构建镜像成功'
echo '>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>正在推送镜像到镜像仓库'
docker push ${IMAGE_ID}
echo '<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<推送镜像到镜像仓库成功'
echo '>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>正在清理本地镜像'
docker rmi ${IMAGE_ID}
echo '<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<清理本地镜像成功'
echo '<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<完成构建服务:'${JOB_NAME}
echo '>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>'
echo '>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>'
echo '>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>'
echo '>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>'
echo '>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>'
echo '>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>'
echo '请拷贝镜像id到下一环节,镜像id为:'
echo ${IMAGE_ID}
echo '<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<'
echo '<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<'
echo '<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<'
echo '<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<'
echo '<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<'
echo '<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<'
```

Dockerfile:

```
FROM registry-vpc.cn-shenzhen.aliyuncs.com/xxx/base_image-master:2790_201911271640
ADD target/*.jar app.jar
ADD startup.sh startup.sh
RUN bash -c 'touch app.jar'
RUN chmod 777 startup.sh
ARG IMAGE_PROJECT_TAG
ENV SW_AGENT_NAME ${IMAGE_PROJECT_TAG}
ENTRYPOINT ["./startup.sh"]
```

# 开发

laas-soa-operate-builder项目作为agent构建器二进制文件部署在服务器

启动时指定参数连接到服务器, 接收服务器的指令和参数进行构建







在服务端设置哪些ip的agent是可以消费哪些事情的, 可以在服务端进行维护agent功能类型表, 也可以由agent启动时指定自己的动作类型, 实现agent订阅动作类型

动作+参数

每隔0.5秒钟请求/comsume_action消费动作, 同时标记该动作id的状态为comsumed(已消费), 返回数据如下

```
[
	{
		"action": "build",
		"action_id": "xxx",
		"action_data": [
			"action": "build",
            "action_id": "xxx",
            "action_data_id": "xxx",
            "project_type": "java",
            "type": "branch", 
            "build_config_file_id": "xxx",
            "build_config_file_version_list": [{"xxx":"xxx"}]
		]
	}
]
```

当对比本地build_config_file_version_list和服务器上的版本不一样时, 携带不一样文件的名称作为参数请求服务器获取数据

```
[
	"data_id": "xxx",
	"param_list": [
		{"build_config_file_version_list": "xxx"}
	]
]
```

响应数据为

```
[
	"data_id": "xxx",
	"data_list":[
		"xxx": ""
	]
]
```

请求/log_data记录日志



可以使用websocket

也可以使用http进






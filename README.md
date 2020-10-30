自己写一个jenkins-构建器/gitlab-runner?



builder(构建器)

基于docker进行资源隔离和构建器管理

运行依赖于服务器

启动时指定服务端的连接信息(基于websocket?), 连接到服务器

服务端通过ssh连接到构建机器进行初始化构建器





构建过程为:

根据项目目录创建并切换到运行目录, 目录结构如下:

```
该action的根目录
	dependency_lib
	git配置名称
		仓库目录
			startup.sh
			cache	
			branch
				master
					source
                    build
				xxx
			tag
				xxx
					source
                    build

```

cache中保存该仓库的构建缓存数据, source中存源码, build中存构建脚本文件, 使用version文件记录构建脚本文件的版本, 当版本一致时则不拉取最新构建脚本, 反之则拉取最新构建脚本

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

​	构建信息:

​			编译镜像名称

​			镜像仓库地址

​			compile_to_execute.sh

​			package_to_docker.sh

​			Dockerfile

​	git服务器信息:

​			连接信息: ip、port、username、password, 该连接账号的权限只需要只读即可

​			是否默认

​	

​	配置项目信息:

​		git服务器: 默认使用默认的git服务器

​		地址: 可从git服务器选择, 也可以直接输入

​		项目描述: 如果是gitlab服务器则自动带出项目描述, 也可以直接输入

​		项目类型: java/php/nodejs/python

​		构建信息: 选择一个构建信息, 可进行预览

​	镜像构建:

​		镜像仓库地址

​		编译镜像名称

​		compile_to_execute.sh

​		package_to_docker.sh

​		Dockerfile



和服务端的维持关系:

使用websocket方式 或者 http

当网络连接状态良好时使用websocket, 当网络连接状态不好时使用http

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

## 服务端

在服务端设置哪些ip的agent是可以消费哪些动作的, 可以在服务端进行维护agent功能类型表

通过配置的agent列表以及他的ssh连接信息, 连接到服务器去执行初始化agent指令, 在启动时设置启动参数: 服务端地址、服务端认证信息、订阅的action列表

接收http请求, 在动作和数据队列中对应关联生成一条数据

当客户端消费一条动作时, 标记该动作id的状态为comsumed(已消费)

## agent

laas-soa-operate-builder项目作为agent构建器二进制文件部署在服务器

启动时指定参数连接到服务器, 接收服务器的指令和参数进行构建

每隔05S请求服务端, 消费自定订阅的动作的动作队列中的数据(适用于这里的场景), 然后按照上面的接口流程去走

可以使用websocket, 也可以使用http



### 消费动作和同步数据的过程

#### 消费动作

请求/comsume_action消费动作, 返回数据如下

```
[
	{
		"action_id": "xxx",
		"action_type": "build",
		"action_data": [
            "cache": {
            	"docker-registry": {"versin": "", "id": "xxx"},
            	"git-repo": {"version": "", "id": "xxx"},
            	"build-script": [{"path": "", "versin": "", "id": "xxx"}]
            },
            "non-cached": {
            	"git-repo-url": "xxx",
            	"docker-registry-url": "xxx",
            	"type": "branch", 
            	"type_value": "master",
            	"project_type": "java",
            },        
		]
		"startup.sh": "",
	}
]
```

同步源码到本地目录

#### 请求同步差异数据

当对比本地build_config_file_version_list和服务器上的版本不一样时, 携带不一样文件的名称作为参数请求服务器获取数据

```
[
	{
			"data_id": "xxx",
			"data_name": "xxx",
			"param_list": [{"path": "xxx"}]
	}
]
```

响应数据为

```
[
	{
		"data_id": "xxx",
		"data_name": "xxx",
        "build-script":[
            {
            	"path": "xxx",
            	"data": "xxx"
            }
        ]
	}
]
```

存储字符串/文件都是要到文件中



#### 记录日志

请求/log_data记录日志



### agent的执行命令

当同步完数据之后就可以执行agent的启动命令

暂时模拟

```
start_action.sh

# 构建
docker run -it --name action_build_1 -v /data:实际目录 maven:3-alpine git配置名称/仓库目录/compile_to_execute.sh

compile_to_execute.sh:
mvn clean package -DskipTests


# 打包
docker run -it --name action_build_1 -v /data:实际目录 docker  git配置名称/仓库目录/package_to_docker.sh

package_to_docker.sh:
cat /root/.m2/dockerregistry-auth |  docker login ${DOCKER_REGISTRY_URL} --username ${DOCKER_REGISTRY_USERNAME} --password-stdin
docker build --build-arg IMAGE_PROJECT_TAG=${IMAGE_PROJECT_TAG} -t ${IMAGE_ID}  .
docker push ${IMAGE_ID}
docker rmi ${IMAGE_ID}

Dockerfile:
FROM registry-vpc.cn-shenzhen.aliyuncs.com/xxx/base_image-master:2790_201911271640
ADD target/*.jar app.jar
ADD startup.sh startup.sh
RUN bash -c 'touch app.jar'
RUN chmod 777 startup.sh
ARG IMAGE_PROJECT_TAG
ENV SW_AGENT_NAME ${IMAGE_PROJECT_TAG}
ENTRYPOINT ["./startup.sh"]

# 修改这一次构建打包目标的镜像id
```


import json
import sys
import time

import requests

# 启动时需要指定的命令行参数
expect_args = {
    "server_uri": "",  # 必须传入, 设置服务端的连接地址, 值例如: --server_uri=http://172.31.42.235:8080
    "business_type_list": []
    # 可以传入, 用以约束当前agent的业务列表, 使用逗号进行分割, 默认使用服务端的设置的业务列表用以订阅, 值例如: --business_type_list=build
}


# 格式化命令行参数
def check_command_line_args():
    if len(sys.argv) < 2:
        raise Exception("请设置服务端地址")
    input_server_uri = sys.argv[1]
    index = input_server_uri.find("=") + 1
    if index > len(input_server_uri) - 1:
        raise Exception("请设置服务端地址")
    input_server_uri = input_server_uri[index:]
    if input_server_uri == "":
        raise Exception("请设置服务端地址")
    expect_args["server_uri"] = input_server_uri
    print("current server uri is: %s" % expect_args["server_uri"])

    if len(sys.argv) > 2:
        input_business_type_list = sys.argv[2]
        index = input_business_type_list.find("=") + 1
        if index < len(input_business_type_list):
            business_type_list = input_business_type_list[index:]
            if business_type_list != "":
                business_type_list = business_type_list.split(",")
                expect_args["business_type_list"] = business_type_list
                print("current business type list is: %s" % str(expect_args["business_type_list"]))


# 请求数据
def request_data(url, data):
    r = requests.post(expect_args["server_uri"] + url, data, {})
    return r.json()


# 消费业务
def consume_business():
    while True:
        resp = request_data("/agent/consume_business",
                            json.dumps({"business_type_list": expect_args["business_type_list"]}))
        print(resp)
        time.sleep(1)


if __name__ == '__main__':
    check_command_line_args()
    consume_business()

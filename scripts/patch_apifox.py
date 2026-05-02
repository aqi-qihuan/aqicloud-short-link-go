#!/usr/bin/env python3
"""
补全 Apifox JSON — 对比 Go 项目实际 API，修复问题并添加缺失接口。

差异分析：
============================================

已有的 API (35个)：全部保留
缺失的 API (6个)：补充

缺失列表：
1. 更新短链     POST /api/link/v1/update     → 短链/短链
2. 短链检查     GET  /api/link/v1/check       → 短链/短链 (RPC 内部)
3. 订单分页     POST /api/order/v1/page       → 商品和支付
4. 查询订单状态 GET  /api/order/v1/query_state → 商品和支付
5. 支付宝回调   POST /api/callback/order/v1/alipay → 商品和支付
6. 流量扣减     POST /api/traffic/v1/reduce   → 账户 (RPC 内部)

修复列表：
1. 解析短链 path: /026m8O3a → /{shortLinkCode}
2. 删除分组 path: /api/group/v1/del/657242090343223296 → /api/group/v1/del/{group_id}
3. 组详情 path: /api/group/v1/detail/657242090343223296 → /api/group/v1/detail/{group_id}
4. 商品详情 path: /api/product/v1/detail/1 → /api/product/v1/detail/{id}
5. 流量包详情 path: /api/traffic/v1/detail/2865 → /api/traffic/v1/detail/{id}
6. 删除短链、高频IP、高频来源、设备信息、分钟趋势 — 补充缺失的 token header
"""

import json
import sys
import copy

INPUT = "/Users/aqi/Documents/GitHub/aqicloud-short-link/flink短链.apifox.json"
OUTPUT = "/Users/aqi/Documents/GitHub/aqicloud-short-link-go/.workbuddy/artifacts/flink短链.apifox.json"

# Helper: generate a unique ID string
_next_id = 337048200
def next_api_id():
    global _next_id
    _next_id += 1
    return str(_next_id)

_next_case_id = 292101920
def next_case_id():
    global _next_case_id
    _next_case_id += 1
    return _next_case_id

# Template for token header
def token_header():
    return [{
        "id": "auto_" + next_api_id(),
        "name": "token",
        "required": True,
        "description": "JWT 登录令牌",
        "example": "{{token}}",
        "type": "string"
    }]

def token_case_header():
    return [{
        "name": "token",
        "value": "{{token}}",
        "enable": True,
        "id": "auto_ch_" + str(next_case_id()),
        "relatedName": "token"
    }]

# Build an API item following Apifox structure
def build_api(name, method, path, description="", ordering=0,
              request_body_type="text/plain", request_body_example="",
              request_params=None, extra_headers=None,
              query_params=None, path_params=None):
    api_id = next_api_id()
    case_id = next_case_id()

    headers = token_header()
    if extra_headers:
        headers.extend(extra_headers)

    params = {
        "path": path_params or [],
        "query": query_params or [],
        "cookie": [],
        "header": headers,
    }

    req_body = {
        "type": request_body_type,
        "parameters": [],
        "oasExtensions": ""
    }
    if request_body_example:
        req_body["examples"] = [{
            "value": request_body_example,
            "mediaType": "text/plain",
            "description": ""
        }]

    case_params = {
        "path": [],
        "query": [],
        "cookie": [],
        "header": token_case_header(),
    }
    if extra_headers:
        for h in extra_headers:
            case_params["header"].append({
                "name": h["name"],
                "value": h.get("example", ""),
                "enable": True,
                "id": "auto_ch2_" + str(case_id),
                "relatedName": h["name"]
            })

    case_body = {
        "parameters": [],
        "type": request_body_type,
        "generateMode": "normal"
    }
    if request_body_example:
        case_body["data"] = request_body_example

    return {
        "name": name,
        "api": {
            "id": api_id,
            "method": method,
            "path": path,
            "parameters": params,
            "auth": {},
            "securityScheme": {},
            "commonParameters": {},
            "responses": [{
                "id": "resp_" + api_id,
                "code": 200,
                "name": "成功",
                "headers": [],
                "jsonSchema": {
                    "type": "object",
                    "properties": {},
                    "x-apifox-orders": []
                },
                "description": "",
                "contentType": "json",
                "mediaType": "",
                "oasExtensions": ""
            }],
            "responseExamples": [],
            "requestBody": req_body,
            "description": description,
            "tags": [],
            "status": "released",
            "serverId": "",
            "operationId": "",
            "sourceUrl": "",
            "ordering": ordering,
            "cases": [{
                "id": case_id,
                "type": "DEBUG_CASE",
                "path": None,
                "name": "成功",
                "responseId": "resp_" + api_id,
                "parameters": case_params,
                "commonParameters": {},
                "requestBody": case_body,
                "auth": {},
                "securityScheme": {},
                "advancedSettings": {"disabledSystemHeaders": {}},
                "requestResult": None,
                "visibility": "INHERITED",
                "moduleId": 6012174,
                "categoryId": 0,
                "tagIds": [],
                "apiTestDataList": [],
                "preProcessors": [],
                "postProcessors": [],
                "inheritPostProcessors": {},
                "inheritPreProcessors": {}
            }],
            "mocks": [],
            "customApiFields": "{}",
            "advancedSettings": {"disabledSystemHeaders": {}},
            "mockScript": {},
            "codeSamples": [],
            "commonResponseStatus": {},
            "responseChildren": [],
            "visibility": "INHERITED",
            "moduleId": 6012174,
            "oasExtensions": "",
            "type": "http",
            "preProcessors": [],
            "postProcessors": [],
            "inheritPostProcessors": {},
            "inheritPreProcessors": {}
        }
    }


def add_token_header_to_api(api_item):
    """Add missing token header to an API item."""
    api = api_item.get("api", {})
    params = api.get("parameters", {})
    headers = params.get("header", [])

    # Check if token header already exists
    has_token = any(h.get("name") == "token" for h in headers)
    if not has_token:
        headers.append({
            "id": "fix_" + next_api_id(),
            "name": "token",
            "required": True,
            "description": "JWT 登录令牌",
            "example": "{{token}}",
            "type": "string"
        })
        params["header"] = headers

    # Also fix case headers
    for case in api.get("cases", []):
        case_headers = case.get("parameters", {}).get("header", [])
        has_token_case = any(h.get("name") == "token" for h in case_headers)
        if not has_token_case:
            case_headers.append({
                "name": "token",
                "value": "{{token}}",
                "enable": True,
                "id": "fix_ch_" + str(next_case_id()),
                "relatedName": "token"
            })
            case["parameters"]["header"] = case_headers


def fix_path_in_api(api_item, new_path):
    """Fix hardcoded path in API."""
    api_item["api"]["path"] = new_path
    # Also add path parameter definition
    path_param_name = new_path.split("{")[-1].rstrip("}") if "{" in new_path else ""
    if path_param_name:
        api_item["api"]["parameters"]["path"] = [{
            "id": "pp_" + next_api_id(),
            "name": path_param_name,
            "required": True,
            "description": "",
            "example": "1",
            "type": "string"
        }]


def main():
    with open(INPUT, "r", encoding="utf-8") as f:
        data = json.load(f)

    root = data["apiCollection"][0]
    # root.items = [短链, 数据, 账户, 商品和支付]

    link_folder = root["items"][0]  # 短链
    data_folder = root["items"][1]  # 数据
    account_folder = root["items"][2]  # 账户
    shop_folder = root["items"][3]  # 商品和支付

    link_api_folder = link_folder["items"][0]  # 短链/短链
    link_group_folder = link_folder["items"][1]  # 短链/分组

    # ============================================================
    # FIX 1: 解析短链 path
    # ============================================================
    for item in link_api_folder["items"]:
        if item["name"] == "解析短链":
            fix_path_in_api(item, "/{shortLinkCode}")
            item["api"]["description"] = "通过短链码获取原始URL，返回302重定向"
            print("FIX: 解析短链 path → /{shortLinkCode}")

    # ============================================================
    # FIX 2: 删除分组 path
    # ============================================================
    for item in link_group_folder["items"]:
        if item["name"] == "根据id删除分组":
            fix_path_in_api(item, "/api/group/v1/del/{group_id}")
            item["api"]["description"] = "根据分组ID删除指定分组"
            print("FIX: 删除分组 path → /api/group/v1/del/{group_id}")

    # ============================================================
    # FIX 3: 组详情 path
    # ============================================================
    for item in link_group_folder["items"]:
        if item["name"] == "组详情":
            fix_path_in_api(item, "/api/group/v1/detail/{group_id}")
            item["api"]["description"] = "获取指定分组的详情信息"
            print("FIX: 组详情 path → /api/group/v1/detail/{group_id}")

    # ============================================================
    # FIX 4: 商品详情 path
    # ============================================================
    for item in shop_folder["items"]:
        if item["name"] == "商品详情":
            fix_path_in_api(item, "/api/product/v1/detail/{id}")
            print("FIX: 商品详情 path → /api/product/v1/detail/{id}")

    # ============================================================
    # FIX 5: 流量包详情 path
    # ============================================================
    for item in account_folder["items"]:
        if "流量包" in item["name"] and "详情" in item["name"]:
            fix_path_in_api(item, "/api/traffic/v1/detail/{id}")
            print("FIX: 流量包详情 path → /api/traffic/v1/detail/{id}")

    # ============================================================
    # FIX 6: 补充缺失的 token header
    # ============================================================
    apis_needing_token_fix = ["删除短链", "高频访问ip", "高频访问来源", "设备信息分布情况", "访问趋势图-分钟分布"]
    for folder_items in [link_api_folder["items"], data_folder["items"]]:
        for item in folder_items:
            if item["name"] in apis_needing_token_fix:
                add_token_header_to_api(item)
                print(f"FIX: {item['name']} — 补充 token header")

    # ============================================================
    # ADD 1: 更新短链 → 短链/短链
    # ============================================================
    update_link = build_api(
        name="更新短链",
        method="post",
        path="/api/link/v1/update",
        description="更新已有短链的标题、原始URL、域名等信息",
        ordering=36,
        request_body_example=json.dumps({
            "id": 658180183031197696,
            "groupId": 658180183031197696,
            "title": "更新后的标题",
            "originalUrl": "https://xdclass.net/#/coursedetail?video_id=2",
            "domain": "g1.fit",
            "domainType": "offical",
            "domainId": 1,
            "code": "0abc0",
            "expired": "2025-12-31"
        }, ensure_ascii=False, indent=1)
    )
    link_api_folder["items"].append(update_link)
    print("ADD: 更新短链 POST /api/link/v1/update")

    # ============================================================
    # ADD 2: 短链检查 → 短链/短链 (RPC内部)
    # ============================================================
    check_link = build_api(
        name="检查短链是否存在",
        method="get",
        path="/api/link/v1/check",
        description="RPC内部接口：检查指定短链码是否存在",
        ordering=42,
        query_params=[{
            "id": "qp_" + next_api_id(),
            "name": "shortLinkCode",
            "required": True,
            "description": "短链压缩码",
            "example": "0abc0",
            "type": "string"
        }],
        request_body_type="none"
    )
    # Remove token header for RPC internal
    check_link["api"]["parameters"]["header"] = []
    check_link["api"]["cases"][0]["parameters"]["header"] = []
    link_api_folder["items"].append(check_link)
    print("ADD: 检查短链是否存在 GET /api/link/v1/check")

    # ============================================================
    # ADD 3: 订单分页 → 商品和支付
    # ============================================================
    order_page = build_api(
        name="订单分页查询",
        method="post",
        path="/api/order/v1/page",
        description="分页查询当前用户的订单列表",
        ordering=30,
        request_body_example=json.dumps({
            "state": "NEW",
            "page": 1,
            "size": 10
        }, ensure_ascii=False, indent=1)
    )
    shop_folder["items"].append(order_page)
    print("ADD: 订单分页查询 POST /api/order/v1/page")

    # ============================================================
    # ADD 4: 查询订单状态 → 商品和支付
    # ============================================================
    order_state = build_api(
        name="查询订单状态",
        method="get",
        path="/api/order/v1/query_state",
        description="查询指定订单的支付状态",
        ordering=36,
        query_params=[{
            "id": "qp_" + next_api_id(),
            "name": "out_trade_no",
            "required": True,
            "description": "订单号",
            "example": "ORDER_202201010001",
            "type": "string"
        }],
        request_body_type="none"
    )
    shop_folder["items"].append(order_state)
    print("ADD: 查询订单状态 GET /api/order/v1/query_state")

    # ============================================================
    # ADD 5: 支付宝回调 → 商品和支付
    # ============================================================
    ali_callback = build_api(
        name="支付宝支付回调",
        method="post",
        path="/api/callback/order/v1/alipay",
        description="支付宝异步支付结果通知回调（由支付宝服务器调用）",
        ordering=42,
        request_body_type="application/x-www-form-urlencoded",
        request_body_example="trade_no=2021101022001412345&out_trade_no=ORDER_202201010001&trade_status=TRADE_SUCCESS&total_amount=2.00&sign=xxx&sign_type=RSA2"
    )
    # 支付宝回调不需要 token
    ali_callback["api"]["parameters"]["header"] = []
    ali_callback["api"]["cases"][0]["parameters"]["header"] = []
    shop_folder["items"].append(ali_callback)
    print("ADD: 支付宝支付回调 POST /api/callback/order/v1/alipay")

    # ============================================================
    # ADD 6: 流量扣减 → 账户 (RPC内部)
    # ============================================================
    traffic_reduce = build_api(
        name="流量扣减",
        method="post",
        path="/api/traffic/v1/reduce",
        description="RPC内部接口：扣减用户当日短链创建流量配额",
        ordering=54,
        request_body_example=json.dumps({
            "accountNo": 1234567890,
            "bizId": "0abc0"
        }, ensure_ascii=False, indent=1),
        extra_headers=[{
            "id": "rpc_" + next_api_id(),
            "name": "rpc-token",
            "required": True,
            "description": "服务间RPC调用令牌",
            "example": "rpc-token-default",
            "type": "string"
        }]
    )
    account_folder["items"].append(traffic_reduce)
    print("ADD: 流量扣减 POST /api/traffic/v1/reduce")

    # ============================================================
    # Add new subfolder for 分组 under 短链 (already exists, just verify)
    # Also add "短链更新" subfolder reference if needed
    # ============================================================

    # ============================================================
    # Update project info
    # ============================================================
    data["info"]["name"] = "flink短链"
    data["info"]["description"] = "AqiCloud 短链接系统 API 文档\n\n" + \
        "架构：Go + Gin + MySQL(分库分表) + Redis + RabbitMQ + Kafka\n" + \
        "微服务：Gateway(:8888), Account(:8001), Data(:8002), Link(:8003), Shop(:8005)\n\n" + \
        "认证方式：JWT Token，通过 header `token` 传递\n\n" + \
        "补充记录 (2026-05-02)：\n" + \
        "- 新增 6 个 API：更新短链、短链检查、订单分页、查询订单状态、支付宝回调、流量扣减\n" + \
        "- 修复 5 个硬编码路径：解析短链、删除分组、组详情、商品详情、流量包详情\n" + \
        "- 修复 5 个 API 缺失的 token header"

    # Write output
    with open(OUTPUT, "w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, indent=2)

    # Count total APIs
    total = 0
    for folder in root["items"]:
        for subfolder in folder.get("items", []):
            total += len(subfolder.get("items", []))
        if not folder.get("items"):
            pass

    print(f"\n✅ 完成！输出: {OUTPUT}")
    print(f"   总 API 数: {total} (原 35 + 新增 6 = 41)")


if __name__ == "__main__":
    main()

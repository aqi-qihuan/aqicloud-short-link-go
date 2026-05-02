#!/usr/bin/env python3
"""
压测数据填充脚本 — 在 MySQL 中预插入短链测试数据

使用方法:
    python3 test/stress/seed_data.py
    python3 test/stress/seed_data.py --count 100000 --host 127.0.0.1

依赖:
    pip install pymysql
"""

import argparse
import hashlib
import random
import string
import sys
import time

try:
    import pymysql
except ImportError:
    print("请安装 pymysql: pip install pymysql")
    sys.exit(1)


# 分库分表配置 (与项目一致)
DB_PREFIX_LIST = ["0", "1", "a"]
TABLE_SUFFIX_LIST = ["0", "a"]  # short_link 分表后缀

DB_CONFIGS = {
    "0": {"database": "aqicloud_link_0"},
    "1": {"database": "aqicloud_link_1"},
    "a": {"database": "aqicloud_link_a"},
}


def base62_encode(num):
    """Base62 编码 (与项目 util.EncodeToBase62 一致)"""
    chars = string.digits + string.ascii_lowercase + string.ascii_uppercase
    if num == 0:
        return chars[0]
    result = []
    while num > 0:
        result.append(chars[num % 62])
        num //= 62
    return ''.join(reversed(result))


def generate_code():
    """生成一个短链码 (模拟项目的分库分表路由)"""
    # 随机 hash 值 -> base62
    hash_val = random.randint(0, 2**32 - 1)
    code = base62_encode(hash_val)[:5]  # 取 5 位
    # 补齐到 5 位
    while len(code) < 5:
        code = '0' + code

    db_prefix = random.choice(DB_PREFIX_LIST)
    table_suffix = random.choice(TABLE_SUFFIX_LIST)
    full_code = db_prefix + code + table_suffix
    return full_code, db_prefix, table_suffix


def md5(text):
    return hashlib.md5(text.encode()).hexdigest().upper()


def main():
    parser = argparse.ArgumentParser(description="Seed test data for stress testing")
    parser.add_argument("--host", default="127.0.0.1", help="MySQL host")
    parser.add_argument("--port", type=int, default=3306, help="MySQL port")
    parser.add_argument("--user", default="root", help="MySQL user")
    parser.add_argument("--password", default="root", help="MySQL password")
    parser.add_argument("--count", type=int, default=10000, help="Number of records to insert")
    parser.add_argument("--batch", type=int, default=500, help="Batch insert size")
    args = parser.parse_args()

    print(f"=== 压测数据填充 ===")
    print(f"MySQL: {args.host}:{args.port}")
    print(f"目标: 插入 {args.count} 条短链记录")
    print()

    # 连接 3 个分库
    connections = {}
    for prefix, config in DB_CONFIGS.items():
        conn = pymysql.connect(
            host=args.host,
            port=args.port,
            user=args.user,
            password=args.password,
            database=config["database"],
            charset="utf8mb4",
        )
        connections[prefix] = conn
        print(f"  已连接: {config['database']}")

    print()

    # 生成唯一 codes
    used_codes = set()
    records = []
    while len(records) < args.count:
        full_code, db_prefix, table_suffix = generate_code()
        if full_code in used_codes:
            continue
        used_codes.add(full_code)

        original_url = f"https://example.com/test-page-{len(records)}/{random.randint(1, 999999)}"
        prefixed_url = f"{random.randint(1000000000000, 9999999999999)}&{original_url}"
        # code 存完整码 (含 db_prefix + table_suffix)，与路由逻辑一致
        # RouteShortLink: first char -> DB, last char -> table
        code_full = full_code

        table_name = f"short_link_{table_suffix}"
        record_id = random.randint(1000000000000000000, 9999999999999999999)

        records.append({
            "db_prefix": db_prefix,
            "table_name": table_name,
            "id": record_id,
            "code": code_full,
            "original_url": prefixed_url,
            "title": f"Perf Test {len(records)}",
            "domain": "g1.fit",
            "sign": md5(original_url),
            "account_no": random.randint(1, 1000),
            "group_id": random.randint(1, 10),
            "state": "ACTIVE",
            "del": 0,
        })

    # 批量插入
    insert_sql = """
        INSERT IGNORE INTO {table} (id, code, original_url, title, domain, sign, account_no, group_id, state, del)
        VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
    """

    total_inserted = 0
    start_time = time.time()

    # 按分库分组
    for db_prefix in DB_CONFIGS:
        db_records = [r for r in records if r["db_prefix"] == db_prefix]
        if not db_records:
            continue

        conn = connections[db_prefix]
        cursor = conn.cursor()

        # 按表分组
        for table_suffix in TABLE_SUFFIX_LIST:
            table_name = f"short_link_{table_suffix}"
            table_records = [r for r in db_records if r["table_name"] == table_name]
            if not table_records:
                continue

            sql = insert_sql.format(table=table_name)
            batch = []
            for r in table_records:
                batch.append((
                    r["id"], r["code"], r["original_url"], r["title"],
                    r["domain"], r["sign"], r["account_no"], r["group_id"],
                    r["state"], r["del"],
                ))

                if len(batch) >= args.batch:
                    cursor.executemany(sql, batch)
                    conn.commit()
                    total_inserted += len(batch)
                    elapsed = time.time() - start_time
                    rate = total_inserted / elapsed if elapsed > 0 else 0
                    print(f"  [{db_prefix}][{table_name}] 已插入 {total_inserted}/{args.count} ({rate:.0f} records/s)")
                    batch = []

            if batch:
                cursor.executemany(sql, batch)
                conn.commit()
                total_inserted += len(batch)
                print(f"  [{db_prefix}][{table_name}] 已插入 {total_inserted}/{args.count}")

        cursor.close()

    elapsed = time.time() - start_time
    print()
    print(f"=== 完成 ===")
    print(f"总插入: {total_inserted} 条, 耗时: {elapsed:.1f}s, 速率: {total_inserted/elapsed:.0f} records/s")

    # 收集测试用的 codes (输出给 k6 脚本使用)
    sample_codes = [full_code for full_code in list(used_codes)[:25]]
    print()
    print(f"样本 codes (复制到 k6 脚本 SHORT_CODES 数组):")
    print(f"  {sample_codes}")

    # 关闭连接
    for conn in connections.values():
        conn.close()


if __name__ == "__main__":
    main()

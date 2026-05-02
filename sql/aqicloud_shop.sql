-- AqiCloud Shop Database
-- Tables: product, product_order (sharded: product_order_0, product_order_1)

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

CREATE DATABASE IF NOT EXISTS `aqicloud_shop` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
USE `aqicloud_shop`;

-- ----------------------------
-- product
-- ----------------------------
DROP TABLE IF EXISTS `product`;
CREATE TABLE `product` (
  `id` bigint NOT NULL,
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '商品标题',
  `detail` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '详情',
  `img` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '图片',
  `level` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '产品层级：FIRST青铜、SECOND黄金、THIRD钻石',
  `old_amount` decimal(16, 0) DEFAULT NULL COMMENT '原价',
  `amount` decimal(16, 2) DEFAULT NULL COMMENT '现价',
  `plugin_type` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '工具类型：SHORT_LINK/QRCODE',
  `day_times` int DEFAULT NULL COMMENT '日次数：短链类型',
  `total_times` int DEFAULT NULL COMMENT '总次数：活码才有',
  `valid_day` int DEFAULT NULL COMMENT '有效天数',
  `gmt_modified` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `gmt_create` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE utf8mb4_bin ROW_FORMAT = Dynamic;

-- Default products
INSERT INTO `product` VALUES (1, '青铜会员-默认', '数据查看支持||日生成短链{{dayTimes}}次||限制跳转50次||默认域名', NULL, 'FIRST', 19, 0.00, 'SHORT_LINK', 2, NULL, 1, '2021-10-14 17:33:44', '2021-10-11 10:49:35');
INSERT INTO `product` VALUES (2, '黄金会员-月度', '数据查看支持||日生成短链{{dayTimes}}次||限制不限制||默认域名', NULL, 'SECOND', 99, 0.01, 'SHORT_LINK', 5, NULL, 30, '2023-09-27 20:49:04', '2021-10-11 10:57:47');
INSERT INTO `product` VALUES (3, '黑金会员-月度', '数据查看支持||日生成短链{{dayTimes}}次||限制不限制||自定义域名', NULL, 'THIRD', 199, 2.00, 'SHORT_LINK', 8, NULL, 30, '2021-10-19 14:36:30', '2021-10-11 11:01:13');

-- ----------------------------
-- product_order_0 (sharded)
-- ----------------------------
DROP TABLE IF EXISTS `product_order_0`;
CREATE TABLE `product_order_0` (
  `id` bigint NOT NULL,
  `product_id` bigint DEFAULT NULL COMMENT '订单类型',
  `product_title` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '商品标题',
  `product_amount` decimal(16, 2) DEFAULT NULL COMMENT '商品单价',
  `product_snapshot` varchar(2048) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '商品快照',
  `buy_num` int DEFAULT NULL COMMENT '购买数量',
  `out_trade_no` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '订单唯一标识',
  `state` varchar(11) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'NEW未支付/PAY已支付/CANCEL超时取消',
  `create_time` datetime DEFAULT NULL COMMENT '订单生成时间',
  `total_amount` decimal(16, 2) DEFAULT NULL COMMENT '订单总金额',
  `pay_amount` decimal(16, 2) DEFAULT NULL COMMENT '订单实际支付价格',
  `pay_type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '支付类型：WECHAT_PAY/ALI_PAY',
  `nickname` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '账号昵称',
  `account_no` bigint DEFAULT NULL COMMENT '用户id',
  `del` int DEFAULT 0 COMMENT '0未删除/1已删除',
  `gmt_modified` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `gmt_create` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `bill_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '发票类型：NO_BILL/电子发票/纸质发票',
  `bill_header` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '发票抬头',
  `bill_content` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '发票内容',
  `bill_receiver_phone` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '发票收票人电话',
  `bill_receiver_email` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '发票收票人邮箱',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_query`(`out_trade_no`, `account_no`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- product_order_1 (sharded)
-- ----------------------------
DROP TABLE IF EXISTS `product_order_1`;
CREATE TABLE `product_order_1` (
  `id` bigint NOT NULL,
  `product_id` bigint DEFAULT NULL COMMENT '订单类型',
  `product_title` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '商品标题',
  `product_amount` decimal(16, 2) DEFAULT NULL COMMENT '商品单价',
  `product_snapshot` varchar(2048) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '商品快照',
  `buy_num` int DEFAULT NULL COMMENT '购买数量',
  `out_trade_no` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '订单唯一标识',
  `state` varchar(11) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'NEW未支付/PAY已支付/CANCEL超时取消',
  `create_time` datetime DEFAULT NULL COMMENT '订单生成时间',
  `total_amount` decimal(16, 2) DEFAULT NULL COMMENT '订单总金额',
  `pay_amount` decimal(16, 2) DEFAULT NULL COMMENT '订单实际支付价格',
  `pay_type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '支付类型：WECHAT_PAY/ALI_PAY',
  `nickname` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '账号昵称',
  `account_no` bigint DEFAULT NULL COMMENT '用户id',
  `del` int DEFAULT 0 COMMENT '0未删除/1已删除',
  `gmt_modified` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `gmt_create` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `bill_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '发票类型：NO_BILL/电子发票/纸质发票',
  `bill_header` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '发票抬头',
  `bill_content` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '发票内容',
  `bill_receiver_phone` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '发票收票人电话',
  `bill_receiver_email` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '发票收票人邮箱',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_query`(`out_trade_no`, `account_no`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

SET FOREIGN_KEY_CHECKS = 1;

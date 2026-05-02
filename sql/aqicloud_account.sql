-- AqiCloud Account Database
-- Tables: account, traffic (sharded: traffic, traffic_0, traffic_1), traffic_task

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

CREATE DATABASE IF NOT EXISTS `aqicloud_account` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
USE `aqicloud_account`;

-- ----------------------------
-- account
-- ----------------------------
DROP TABLE IF EXISTS `account`;
CREATE TABLE `account` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `account_no` bigint DEFAULT NULL,
  `head_img` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '头像',
  `phone` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '手机号',
  `pwd` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '密码',
  `secret` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '盐，用于个人敏感信息处理',
  `mail` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '邮箱',
  `username` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '用户名',
  `auth` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '认证级别：DEFAULT/REALNAME/ENTERPRISE',
  `gmt_create` datetime DEFAULT CURRENT_TIMESTAMP,
  `gmt_modified` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_phone`(`phone`) USING BTREE,
  UNIQUE INDEX `uk_account`(`account_no`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 9 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_bin ROW_FORMAT = Dynamic;

-- ----------------------------
-- traffic (unsharded fallback)
-- ----------------------------
DROP TABLE IF EXISTS `traffic`;
CREATE TABLE `traffic` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `day_limit` int DEFAULT NULL COMMENT '每天限制多少条，短链',
  `day_used` int DEFAULT NULL COMMENT '当天用了多少条，短链',
  `total_limit` int DEFAULT NULL COMMENT '总次数，活码才用',
  `account_no` bigint DEFAULT NULL COMMENT '账号',
  `out_trade_no` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '订单号',
  `level` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '产品层级：FIRST青铜、SECOND黄金、THIRD钻石',
  `expired_date` date DEFAULT NULL COMMENT '过期日期',
  `plugin_type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '插件类型',
  `product_id` bigint DEFAULT NULL COMMENT '商品主键',
  `gmt_create` datetime DEFAULT CURRENT_TIMESTAMP,
  `gmt_modified` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_trade_no`(`out_trade_no`, `account_no`) USING BTREE,
  INDEX `idx_account_no`(`account_no`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_bin ROW_FORMAT = Dynamic;

-- ----------------------------
-- traffic_0 (sharded by account_no % 2 == 0)
-- ----------------------------
DROP TABLE IF EXISTS `traffic_0`;
CREATE TABLE `traffic_0` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `day_limit` int DEFAULT NULL COMMENT '每天限制多少条，短链',
  `day_used` int DEFAULT NULL COMMENT '当天用了多少条，短链',
  `total_limit` int DEFAULT NULL COMMENT '总次数，活码才用',
  `account_no` bigint DEFAULT NULL COMMENT '账号',
  `out_trade_no` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '订单号',
  `level` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '产品层级：FIRST青铜、SECOND黄金、THIRD钻石',
  `expired_date` date DEFAULT NULL COMMENT '过期日期',
  `plugin_type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '插件类型',
  `product_id` bigint DEFAULT NULL COMMENT '商品主键',
  `gmt_create` datetime DEFAULT CURRENT_TIMESTAMP,
  `gmt_modified` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_trade_no`(`out_trade_no`, `account_no`) USING BTREE,
  INDEX `idx_account_no`(`account_no`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_bin ROW_FORMAT = Dynamic;

-- ----------------------------
-- traffic_1 (sharded by account_no % 2 == 1)
-- ----------------------------
DROP TABLE IF EXISTS `traffic_1`;
CREATE TABLE `traffic_1` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `day_limit` int DEFAULT NULL COMMENT '每天限制多少条，短链',
  `day_used` int DEFAULT NULL COMMENT '当天用了多少条，短链',
  `total_limit` int DEFAULT NULL COMMENT '总次数，活码才用',
  `account_no` bigint DEFAULT NULL COMMENT '账号',
  `out_trade_no` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '订单号',
  `level` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '产品层级：FIRST青铜、SECOND黄金、THIRD钻石',
  `expired_date` date DEFAULT NULL COMMENT '过期日期',
  `plugin_type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '插件类型',
  `product_id` bigint DEFAULT NULL COMMENT '商品主键',
  `gmt_create` datetime DEFAULT CURRENT_TIMESTAMP,
  `gmt_modified` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_trade_no`(`out_trade_no`, `account_no`) USING BTREE,
  INDEX `idx_account_no`(`account_no`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_bin ROW_FORMAT = Dynamic;

-- ----------------------------
-- traffic_task
-- ----------------------------
DROP TABLE IF EXISTS `traffic_task`;
CREATE TABLE `traffic_task` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `account_no` bigint DEFAULT NULL,
  `traffic_id` bigint DEFAULT NULL,
  `use_times` int DEFAULT NULL,
  `lock_state` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '锁定状态：LOCK/FINISH/CANCEL',
  `biz_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL COMMENT '唯一标识',
  `gmt_create` datetime DEFAULT CURRENT_TIMESTAMP,
  `gmt_modified` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_biz_id`(`biz_id`) USING BTREE,
  INDEX `idx_release`(`account_no`, `id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_bin ROW_FORMAT = Dynamic;

SET FOREIGN_KEY_CHECKS = 1;

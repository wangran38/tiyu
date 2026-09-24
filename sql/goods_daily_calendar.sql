-- ============================================================
-- goods_daily_calendar 日历价格库存表
-- 用途：酒店(HOTEL)/景区(TOUR)等存在"平周末价差"的业态使用
-- 关联：goods_product.id = product_id, goods_sku.id = sku_id
-- 说明：
--   1. 当 goods_product.stock_type=2（每日动态库存）时启用本表
--   2. 按 (product_id, sku_id, date) 三元组唯一
--   3. sku_id=0 表示商品维度的日历（无规格商品）；>0 表示具体 SKU 日历
--   4. 未命中某天的记录时，价/库存回落到 goods_product.selling_price / goods_sku.stock
--   5. status=0 表示当日停售/闭店/闭园，C 端不可下单
-- ============================================================

DROP TABLE IF EXISTS `goods_daily_calendar`;
CREATE TABLE `goods_daily_calendar` (
  `id`         BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `product_id` BIGINT       NOT NULL DEFAULT 0     COMMENT '关联团购产品ID',
  `sku_id`     BIGINT       NOT NULL DEFAULT 0     COMMENT '关联SKU ID，0表示商品维度',
  `date`       DATE         NOT NULL              COMMENT '日期，仅日期部分有效',
  `price`      DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '当日售价（元）',
  `stock`      INT          NOT NULL DEFAULT 0    COMMENT '当日库存',
  `status`     INT          NOT NULL DEFAULT 1    COMMENT '状态: 1-可售, 0-停售/闭店/闭园',
  `createtime` INT          NOT NULL DEFAULT 0    COMMENT '创建时间（时间戳）',
  `updatetime` INT          NOT NULL DEFAULT 0    COMMENT '更新时间（时间戳）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_product_sku_date` (`product_id`, `sku_id`, `date`),
  KEY `idx_product_id` (`product_id`),
  KEY `idx_sku_id`     (`sku_id`),
  KEY `idx_date`       (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='日历价格库存表（酒店/景区等平周末价差业态使用）';

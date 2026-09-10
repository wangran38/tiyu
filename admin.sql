/*
Navicat MySQL Data Transfer

Source Server         : 本地测试
Source Server Version : 50726
Source Host           : 127.0.0.1:3306
Source Database       : ginstudy

Target Server Type    : MYSQL
Target Server Version : 50726
File Encoding         : 65001

Date: 2022-04-27 12:53:46
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for admin
-- ----------------------------
DROP TABLE IF EXISTS `admin`;
CREATE TABLE `admin` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `username` varchar(200) DEFAULT NULL,
  `nickname` varchar(200) DEFAULT NULL,
  `salt` varchar(200) DEFAULT NULL,
  `age` int(2) DEFAULT NULL,
  `avatar` text,
  `loginfailure` int(10) DEFAULT NULL,
  `logintime` int(10) DEFAULT NULL,
  `loginip` varchar(200) DEFAULT NULL,
  `token` varchar(59) DEFAULT NULL,
  `password` varchar(200) DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of admin
-- ----------------------------
INSERT INTO `admin` VALUES ('1', 'admin', '', 'MRq2', null, null, null, null, null, null, '052931db8810edf6dc84d8c9a908e0d6', null, null);

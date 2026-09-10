-- phpMyAdmin SQL Dump
-- version 4.8.5
-- https://www.phpmyadmin.net/
--
-- 主机： localhost
-- 生成日期： 2022-07-27 18:02:48
-- 服务器版本： 5.7.26
-- PHP 版本： 7.3.4

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET AUTOCOMMIT = 0;
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- 数据库： `ginstudy`
--

-- --------------------------------------------------------

--
-- 表的结构 `auth_rule`
--

CREATE TABLE `auth_rule` (
  `id` bigint(20) NOT NULL,
  `pid` int(11) DEFAULT NULL,
  `type` varchar(255) DEFAULT NULL,
  `icon` varchar(255) DEFAULT NULL,
  `pathname` varchar(255) DEFAULT NULL,
  `title` varchar(255) DEFAULT NULL,
  `remark` varchar(255) DEFAULT NULL,
  `ismenu` tinyint(4) NOT NULL DEFAULT '1' COMMENT '是否启用 默认1 菜单 0 文件',
  `created` int(11) DEFAULT NULL,
  `updated` int(11) DEFAULT NULL,
  `deletetime` int(11) DEFAULT NULL,
  `weigh` int(11) DEFAULT NULL,
  `status` varchar(40) DEFAULT NULL,
  `component` varchar(255) DEFAULT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8;

--
-- 转存表中的数据 `auth_rule`
--

INSERT INTO `auth_rule` (`id`, `pid`, `type`, `icon`, `pathname`, `title`, `remark`, `ismenu`, `created`, `updated`, `deletetime`, `weigh`, `status`, `component`) VALUES
(2, 1, 'file', 'el-icon-notebook-2', '/system/admin', '用户列表', '用户列表', 0, 1614759419, NULL, NULL, 1, 'normal', 'system/admin/index'),
(3, 1, 'file', 'el-icon-notebook-2', '/system/rules', '前端菜单', '前端菜单', 0, 1614759461, NULL, NULL, 2, 'normal', 'system/rules/index'),
(4, 1, 'file', 'el-icon-notebook-2', '/system/group', '管理组别', '管理组别', 0, 1614759515, NULL, NULL, 3, 'normal', 'system/group'),
(1, 0, 'file', 'el-icon-menu', '/system', '系统管理', '用户管理', 1, 1621413412, NULL, NULL, 0, 'normal', NULL),
(5, 0, 'file', 'el-icon-menu', '/personalPerformance', '公共配置', '公共配置', 1, 1619409666, NULL, NULL, 0, 'normal', NULL),
(226, 225, 'file', 'el-icon-notebook-2', '/personalPerformance/allStaffEvaluation', '全员民主测评', '全员民主测评', 0, 1619415264, NULL, NULL, 0, 'normal', NULL),
(6, 0, 'file', 'el-icon-menu', '/market', '市场管理', '市场管理', 1, 1620270606, NULL, NULL, 0, 'normal', NULL),
(236, 235, 'file', 'el-icon-tickets', '/judicature/petitionInvolLitigation', '涉诉信访统计', '涉诉信访统计', 0, 1620270690, NULL, NULL, 60, 'normal', NULL),
(237, 235, 'file', 'el-icon-tickets', '/judicature/caseQuality', '案件质量评查', '案件质量评查', 0, 1620270741, NULL, NULL, 50, 'normal', NULL),
(238, 235, 'file', 'el-icon-tickets', '/judicature/caseFiling', '案件归档登记', '案件归档登记', 0, 1620270798, NULL, NULL, 40, 'normal', NULL),
(243, 0, 'file', 'el-icon-menu', '/postGoal', '代理商管理', '代理商管理', 1, 1620442073, NULL, NULL, 0, 'normal', NULL),
(245, 243, 'file', 'el-icon-notebook-2', '/postGoal/jobTarget', '岗位目标', '岗位目标', 0, 1620442243, NULL, NULL, 0, 'normal', NULL),
(246, 243, 'file', 'el-icon-notebook-2', '/postGoal/personnelTasks', '人员任职', '人员任职', 0, 1620442339, NULL, NULL, 0, 'normal', NULL);

--
-- 转储表的索引
--

--
-- 表的索引 `auth_rule`
--
ALTER TABLE `auth_rule`
  ADD PRIMARY KEY (`id`);

--
-- 在导出的表使用AUTO_INCREMENT
--

--
-- 使用表AUTO_INCREMENT `auth_rule`
--
ALTER TABLE `auth_rule`
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=257;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
